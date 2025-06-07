package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/launchstack/backend/config"
	"github.com/launchstack/backend/models"
	"github.com/sirupsen/logrus"
)

// DockerClient is an interface that abstracts Docker operations
type DockerClient interface {
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *string, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error)
	ContainerStats(ctx context.Context, containerID string, stream bool) (types.ContainerStats, error)
	ImagePull(ctx context.Context, refStr string, options types.ImagePullOptions) (io.ReadCloser, error)
	NetworkInspect(ctx context.Context, networkID string, options types.NetworkInspectOptions) (types.NetworkResource, error)
}

// Manager handles Docker container operations
type Manager struct {
	client     DockerClient
	config     *config.Config
	logger     *logrus.Logger
	caddyMgr   *CaddyManager
}

// CaddyManager handles Caddy configuration updates
type CaddyManager struct {
	caddyFilePath string
	reloadCmd     string
}

// NewDockerClient creates a new Docker client
func NewDockerClient(host string) (DockerClient, error) {
	// Always use the Docker API endpoint
	dockerHost := "http://10.1.1.81:2375"
	os.Setenv("DOCKER_HOST", dockerHost)
	
	return client.NewClientWithOpts(
		client.WithHost(dockerHost),
		client.WithAPIVersionNegotiation(),
	)
}

// NewManager creates a new container manager
func NewManager(client DockerClient, cfg *config.Config, logger *logrus.Logger) *Manager {
	caddyMgr := &CaddyManager{
		caddyFilePath: "/etc/caddy/Caddyfile",
		reloadCmd:     "caddy reload --config /etc/caddy/Caddyfile",
	}
	
	return &Manager{
		client:     client,
		config:     cfg,
		logger:     logger,
		caddyMgr:   caddyMgr,
	}
}

// generateContainerName creates a unique container name for a user instance
func generateContainerName(userID uuid.UUID, instanceName string) string {
	// Remove any spaces and special characters from the instance name
	sanitizedName := strings.ToLower(strings.ReplaceAll(instanceName, " ", "-"))
	sanitizedName = strings.ReplaceAll(sanitizedName, "_", "-")
	
	// Create a unique identifier by combining user ID (first 8 chars) and sanitized name
	return fmt.Sprintf("n8n-%s-%s", userID.String()[:8], sanitizedName)
}

// generateSubdomain creates a unique subdomain for the instance
func generateSubdomain(userID uuid.UUID, instanceName string) string {
	// Remove any spaces and special characters from the instance name
	sanitizedName := strings.ToLower(strings.ReplaceAll(instanceName, " ", "-"))
	sanitizedName = strings.ReplaceAll(sanitizedName, "_", "-")
	
	// Create a unique identifier by combining user ID (first 8 chars) and sanitized name
	return fmt.Sprintf("%s-%s", sanitizedName, userID.String()[:8])
}

// generateDataDir creates the data directory for an instance
func (m *Manager) generateDataDir(containerName string) string {
	return filepath.Join(m.config.N8N.DataDir, containerName)
}

// CreateInstance creates a new n8n instance
func (m *Manager) CreateInstance(ctx context.Context, user models.User, instanceReq models.Instance) (*models.Instance, error) {
	// Check if user has reached their instance limit
	instancesLimit := user.GetInstancesLimit()
	if instancesLimit <= 0 {
		return nil, fmt.Errorf("user has no instance allocation")
	}
	
	// TODO: Check how many instances the user already has
	
	// Generate container name and subdomain
	containerName := generateContainerName(user.ID, instanceReq.Name)
	subdomain := generateSubdomain(user.ID, instanceReq.Name)
	
	// Create instance record
	instance := &models.Instance{
		UserID:       user.ID,
		Name:         instanceReq.Name,
		Description:  instanceReq.Description,
		Status:       models.StatusPending,
		Host:         subdomain,
		URL:          fmt.Sprintf("%s.%s", subdomain, m.config.Server.Domain),
		CPULimit:     user.GetCPULimit(),
		MemoryLimit:  user.GetMemoryLimit(),
		StorageLimit: user.GetStorageLimit(),
	}
	
	// Create host bind mount directories
	dataDir := m.generateDataDir(containerName)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	
	filesDir := filepath.Join(dataDir, "files")
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create files directory: %w", err)
	}
	
	// Assign a port for the container
	port, err := m.findAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("failed to find available port: %w", err)
	}
	instance.Port = port
	
	// Pull the latest n8n image
	reader, err := m.client.ImagePull(ctx, m.config.N8N.BaseImage, types.ImagePullOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()
	io.Copy(io.Discard, reader) // Discard the output
	
	// Set up environment variables for the container
	env := []string{
		"NODE_ENV=production",
		fmt.Sprintf("N8N_HOST=%s", instance.URL),
		"N8N_PROTOCOL=https",
		fmt.Sprintf("WEBHOOK_URL=https://%s", instance.URL),
		"N8N_BASIC_AUTH_ACTIVE=true",
		fmt.Sprintf("N8N_BASIC_AUTH_USER=%s", subdomain),
		fmt.Sprintf("N8N_BASIC_AUTH_PASSWORD=%s", uuid.New().String()[:8]),
	}
	
	// Set up container CPU and memory limits
	cpuLimit := int64(instance.CPULimit * 1000000) // Convert to CPU shares
	memoryLimit := int64(instance.MemoryLimit * 1024 * 1024) // Convert MB to bytes
	
	// Create the container
	resp, err := m.client.ContainerCreate(
		ctx,
		&container.Config{
			Image: m.config.N8N.BaseImage,
			Env:   env,
			User:  "root", // Run as root to ensure permission for host bind mounts
			Labels: map[string]string{
				"com.launchstack.instance.id":   instance.ID.String(),
				"com.launchstack.user.id":       user.ID.String(),
				"com.launchstack.managed":       "true",
				"com.centurylinklabs.watchtower.enable": "true",
			},
		},
		&container.HostConfig{
			RestartPolicy: container.RestartPolicy{
				Name: "always",
			},
			Resources: container.Resources{
				Memory:   memoryLimit,
				NanoCPUs: cpuLimit,
			},
			Binds: []string{
				fmt.Sprintf("%s:/home/node/.n8n", dataDir),
				fmt.Sprintf("%s:/files", filesDir),
			},
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				m.config.Docker.Network: {
					NetworkID: m.config.Docker.Network,
				},
			},
		},
		nil,
		containerName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	
	// Update container ID in the instance
	instance.ContainerID = resp.ID
	
	// Start the container
	if err := m.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}
	
	// Update instance status
	instance.Status = models.StatusRunning
	
	// Update Caddy configuration
	if err := m.updateCaddyConfig(instance); err != nil {
		m.logger.WithError(err).Error("Failed to update Caddy configuration")
		// We still return the instance even if Caddy update fails
		// as the container is running
	}
	
	return instance, nil
}

// findAvailablePort finds an available port in the configured range
func (m *Manager) findAvailablePort() (int, error) {
	// TODO: Implement a proper port allocation strategy
	// For now, we'll just return the start of the range
	return m.config.N8N.PortRangeStart, nil
}

// updateCaddyConfig updates the Caddy configuration to add the new instance
func (m *Manager) updateCaddyConfig(instance *models.Instance) error {
	// Read the current Caddyfile
	caddyFile, err := os.ReadFile(m.caddyMgr.caddyFilePath)
	if err != nil {
		return fmt.Errorf("failed to read Caddyfile: %w", err)
	}
	
	// Create the new configuration block for this instance
	newBlock := fmt.Sprintf(`
%s {
	reverse_proxy http://127.0.0.1:%d
	tls {
		dns cloudflare {env.CF_API_TOKEN}
	}
	encode gzip
}
`, instance.URL, instance.Port)
	
	// Append the new block to the Caddyfile
	newCaddyFile := string(caddyFile) + newBlock
	
	// Write the updated Caddyfile
	if err := os.WriteFile(m.caddyMgr.caddyFilePath, []byte(newCaddyFile), 0644); err != nil {
		return fmt.Errorf("failed to write Caddyfile: %w", err)
	}
	
	// Reload Caddy
	// Execute the reload command
	cmd := exec.Command("sh", "-c", m.caddyMgr.reloadCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reload Caddy: %s, error: %w", string(output), err)
	}
	
	return nil
}

// StopInstance stops a running instance
func (m *Manager) StopInstance(ctx context.Context, instanceID uuid.UUID) error {
	// Get the instance from the database
	// instance, err := m.db.GetInstance(instanceID)
	// if err != nil {
	//     return fmt.Errorf("failed to get instance: %w", err)
	// }
	
	// For now, we'll assume we have the container ID
	containerID := "dummy-container-id"
	
	// Stop the container
	timeout := 30 * time.Second
	if err := m.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	
	// Update instance status
	// instance.Status = models.StatusStopped
	// if err := m.db.UpdateInstance(instance); err != nil {
	//     return fmt.Errorf("failed to update instance: %w", err)
	// }
	
	return nil
}

// StartInstance starts a stopped instance
func (m *Manager) StartInstance(ctx context.Context, instanceID uuid.UUID) error {
	// Get the instance from the database
	// instance, err := m.db.GetInstance(instanceID)
	// if err != nil {
	//     return fmt.Errorf("failed to get instance: %w", err)
	// }
	
	// For now, we'll assume we have the container ID
	containerID := "dummy-container-id"
	
	// Start the container
	if err := m.client.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}
	
	// Update instance status
	// instance.Status = models.StatusRunning
	// if err := m.db.UpdateInstance(instance); err != nil {
	//     return fmt.Errorf("failed to update instance: %w", err)
	// }
	
	return nil
}

// DeleteInstance removes an instance
func (m *Manager) DeleteInstance(ctx context.Context, instanceID uuid.UUID) error {
	// Get the instance from the database
	// instance, err := m.db.GetInstance(instanceID)
	// if err != nil {
	//     return fmt.Errorf("failed to get instance: %w", err)
	// }
	
	// For now, we'll assume we have the container ID
	containerID := "dummy-container-id"
	
	// Stop the container if it's running
	timeout := 30 * time.Second
	_ = m.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
	
	// Remove the container
	if err := m.client.ContainerRemove(ctx, containerID, container.RemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}
	
	// TODO: Remove host bind mount directory
	
	// TODO: Update Caddy configuration to remove this instance
	
	// Update instance status
	// instance.Status = models.StatusDeleted
	// if err := m.db.UpdateInstance(instance); err != nil {
	//     return fmt.Errorf("failed to update instance: %w", err)
	// }
	
	return nil
}

// GetInstanceStats retrieves resource usage stats for an instance
func (m *Manager) GetInstanceStats(ctx context.Context, containerID string) (*models.ResourceUsage, error) {
	// Get container stats
	stats, err := m.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer stats.Body.Close()
	
	// Parse the stats
	var statsJSON types.StatsJSON
	if err := json.NewDecoder(stats.Body).Decode(&statsJSON); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}
	
	// Calculate CPU usage percentage
	cpuDelta := float64(statsJSON.CPUStats.CPUUsage.TotalUsage - statsJSON.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(statsJSON.CPUStats.SystemUsage - statsJSON.PreCPUStats.SystemUsage)
	cpuUsage := 0.0
	if systemDelta > 0 && cpuDelta > 0 {
		cpuUsage = (cpuDelta / systemDelta) * float64(len(statsJSON.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	
	// Calculate memory usage
	memoryUsage := statsJSON.MemoryStats.Usage
	
	// Create resource usage record
	usage := &models.ResourceUsage{
		Timestamp:   time.Now(),
		CPUUsage:    cpuUsage,
		MemoryUsage: int64(memoryUsage),
		DiskUsage:   0, // TODO: Implement disk usage calculation
		NetworkIn:   int64(statsJSON.Networks["eth0"].RxBytes),
		NetworkOut:  int64(statsJSON.Networks["eth0"].TxBytes),
	}
	
	return usage, nil
}

// ListInstances lists all containers managed by LaunchStack
func (m *Manager) ListInstances(ctx context.Context) ([]types.Container, error) {
	filters := filters.NewArgs()
	filters.Add("label", "com.launchstack.managed=true")
	
	return m.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters,
	})
}

// GetInstanceByID retrieves a container by instance ID
func (m *Manager) GetInstanceByID(ctx context.Context, instanceID uuid.UUID) (types.Container, error) {
	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("com.launchstack.instance.id=%s", instanceID.String()))
	
	containers, err := m.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters,
	})
	if err != nil {
		return types.Container{}, err
	}
	
	if len(containers) == 0 {
		return types.Container{}, fmt.Errorf("container not found for instance ID: %s", instanceID)
	}
	
	return containers[0], nil
}

// GetInstancesByUserID retrieves all containers for a user
func (m *Manager) GetInstancesByUserID(ctx context.Context, userID uuid.UUID) ([]types.Container, error) {
	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("com.launchstack.user.id=%s", userID.String()))
	
	return m.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters,
	})
} 