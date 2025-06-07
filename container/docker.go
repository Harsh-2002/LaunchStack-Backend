package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/launchstack/backend/config"
	"github.com/launchstack/backend/db"
	"github.com/launchstack/backend/models"
	"github.com/sirupsen/logrus"
)

// DockerClient defines the interface for Docker operations
type DockerClient interface {
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform interface{}, containerName string) (container.ContainerCreateCreatedBody, error)
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error
	ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerStats(ctx context.Context, containerID string, stream bool) (types.ContainerStats, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ImagePull(ctx context.Context, refStr string, options types.ImagePullOptions) (io.ReadCloser, error)
	NetworkInspect(ctx context.Context, networkID string, options types.NetworkInspectOptions) (types.NetworkResource, error)
}

// DockerClientWrapper wraps the Docker client to implement our interface
type DockerClientWrapper struct {
	*client.Client
}

// ContainerCreate wraps the Docker client ContainerCreate method
func (d *DockerClientWrapper) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform interface{}, containerName string) (container.ContainerCreateCreatedBody, error) {
	return d.Client.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, containerName)
}

// ContainerInspect wraps the Docker client ContainerInspect method
func (d *DockerClientWrapper) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return d.Client.ContainerInspect(ctx, containerID)
}

// DockerManager handles Docker container operations
type DockerManager struct {
	client     DockerClient
	config     *config.Config
	logger     *logrus.Logger
	dnsManager *DNSManager
}

// NewDockerClient creates a new Docker client
func NewDockerClient(host string) (DockerClient, error) {
	// Always use the Docker API endpoint
	dockerHost := "http://10.1.1.81:2375"
	os.Setenv("DOCKER_HOST", dockerHost)
	
	c, err := client.NewClientWithOpts(
		client.WithHost(dockerHost),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}
	
	return &DockerClientWrapper{Client: c}, nil
}

// NewManager creates a new Docker container manager
func NewManager(client DockerClient, cfg *config.Config, logger *logrus.Logger) Manager {
	// Create a DNS manager
	dnsManager := NewDNSManager(logger)
	
	return &DockerManager{
		client:     client,
		config:     cfg,
		logger:     logger,
		dnsManager: dnsManager,
	}
}

// Using shared implementation from shared.go

// generateDataDir creates the data directory for an instance
func (m *DockerManager) generateDataDir(containerName string) string {
	return filepath.Join(m.config.N8N.DataDir, containerName)
}

// CreateInstance creates a new n8n instance
func (m *DockerManager) CreateInstance(ctx context.Context, user models.User, instanceReq models.Instance) (*models.Instance, error) {
	// Check if user has reached their instance limit
	instancesLimit := user.GetInstancesLimit()
	if instancesLimit <= 0 {
		return nil, fmt.Errorf("user has no instance allocation")
	}
	
	// TODO: Check how many instances the user already has
	
	// Generate container name and subdomain
	containerName := GenerateContainerName(user.ID, instanceReq.Name)
	subdomain := GenerateEasySubdomain(containerName)
	
	// Create instance record
	instance := &models.Instance{
		UserID:       user.ID,
		Name:         instanceReq.Name,
		Description:  instanceReq.Description,
		Status:       models.StatusPending,
		Host:         subdomain,
		URL:          fmt.Sprintf("%s.%s", subdomain, m.config.Server.Domain),
		// We don't use CPU limits because of cgroup issues, but still store the value
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
	
	// Create data directory permissions
	if err := os.Chmod(dataDir, 0777); err != nil {
		return nil, fmt.Errorf("failed to set data directory permissions: %w", err)
	}
	
	// Set up container memory limits (CPU limits disabled due to cgroup issues)
	memoryLimit := int64(instance.MemoryLimit * 1024 * 1024) // Convert MB to bytes
	
	// Create host config with optional resource limits
	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		Binds: []string{
			fmt.Sprintf("%s:/home/node/.n8n", dataDir),
			fmt.Sprintf("%s:/files", filesDir),
		},
	}
	
	// Only add memory limits if non-zero
	if memoryLimit > 0 {
		hostConfig.Resources.Memory = memoryLimit
	}
	
	// Pull the latest n8n image
	m.logger.Debug("Pulling the latest n8n image")
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
	
	// Create the container
	m.logger.WithFields(logrus.Fields{
		"image":      m.config.N8N.BaseImage,
		"network":    m.config.Docker.Network,
		"subnet":     m.config.Docker.NetworkSubnet,
		"memory_mb":  instance.MemoryLimit,
	}).Debug("Creating Docker container")

	resp, err := m.client.ContainerCreate(
		ctx,
		&container.Config{
			Image: m.config.N8N.BaseImage,
			Env:   env,
			User:  "root", // Run as root to ensure permission for host bind mounts
			// Expose the default n8n port (5678)
			ExposedPorts: map[nat.Port]struct{}{
				nat.Port("5678/tcp"): {},
			},
			Labels: map[string]string{
				"com.launchstack.instance.id":   instance.ID.String(),
				"com.launchstack.user.id":       user.ID.String(),
				"com.launchstack.managed":       "true",
				"com.centurylinklabs.watchtower.enable": "true",
			},
		},
		hostConfig,
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
		m.logger.WithError(err).Error("Failed to create container")
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	
	// Update container ID in the instance
	instance.ContainerID = resp.ID
	m.logger.WithField("container_id", resp.ID).Info("Container created successfully")
	
	// Start the container
	m.logger.WithField("container_id", resp.ID).Debug("Starting container")
	if err := m.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		m.logger.WithError(err).Error("Failed to start container")
		return nil, fmt.Errorf("failed to start container: %w", err)
	}
	m.logger.WithField("container_id", resp.ID).Info("Container started successfully")
	
	// Get the container's IP address
	container, err := m.client.ContainerInspect(ctx, resp.ID)
	if err != nil {
		m.logger.WithError(err).Error("Failed to inspect container")
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}
	
	// Get the container's IP address in the n8n network
	containerIP := container.NetworkSettings.Networks[m.config.Docker.Network].IPAddress
	if containerIP == "" {
		m.logger.Error("Container IP address not found")
		return nil, fmt.Errorf("container IP address not found")
	}
	
	// Create single DNS record for the container: {subdomain}.docker -> Container IP
	dockerDNS := fmt.Sprintf("%s.docker", subdomain)
	
	// Add DNS record to AdGuard
	if err := m.dnsManager.AddDNSRewrite(dockerDNS, containerIP); err != nil {
		m.logger.WithError(err).Error("Failed to add DNS record for Docker name")
		// Non-fatal error, continue
	}
	
	m.logger.WithFields(logrus.Fields{
		"domain": dockerDNS,
		"ip":     containerIP,
	}).Info("Created DNS record for container")
	
	// Update instance status
	instance.Status = models.StatusRunning
	
	return instance, nil
}

// StopInstance stops an instance
func (m *DockerManager) StopInstance(ctx context.Context, instanceID uuid.UUID) error {
	// Get the instance from the database
	instance, err := db.GetInstanceByID(instanceID)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}
	
	// Make sure we have a container ID
	if instance.ContainerID == "" {
		return fmt.Errorf("instance has no container ID")
	}
	
	m.logger.WithFields(logrus.Fields{
		"instance_id":  instance.ID,
		"container_id": instance.ContainerID,
	}).Info("Stopping container")
	
	// Stop the container
	timeout := 30 * time.Second
	if err := m.client.ContainerStop(ctx, instance.ContainerID, &timeout); err != nil {
		m.logger.WithError(err).Error("Failed to stop container")
		return fmt.Errorf("failed to stop container: %w", err)
	}
	
	// Update instance status
	instance.Status = models.StatusStopped
	if err := db.UpdateInstance(instance); err != nil {
		m.logger.WithError(err).Warn("Failed to update instance status")
		// We still return success since the container was stopped
	}
	
	m.logger.WithField("instance_id", instance.ID).Info("Container stopped successfully")
	return nil
}

// StartInstance starts an instance
func (m *DockerManager) StartInstance(ctx context.Context, instanceID uuid.UUID) error {
	// Get the instance from the database
	instance, err := db.GetInstanceByID(instanceID)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}
	
	// Make sure we have a container ID
	if instance.ContainerID == "" {
		return fmt.Errorf("instance has no container ID")
	}
	
	m.logger.WithFields(logrus.Fields{
		"instance_id":  instance.ID,
		"container_id": instance.ContainerID,
	}).Info("Starting container")
	
	// Start the container
	if err := m.client.ContainerStart(ctx, instance.ContainerID, types.ContainerStartOptions{}); err != nil {
		m.logger.WithError(err).Error("Failed to start container")
		return fmt.Errorf("failed to start container: %w", err)
	}
	
	// Update instance status
	instance.Status = models.StatusRunning
	if err := db.UpdateInstance(instance); err != nil {
		m.logger.WithError(err).Warn("Failed to update instance status")
		// We still return success since the container was started
	}
	
	m.logger.WithField("instance_id", instance.ID).Info("Container started successfully")
	return nil
}

// DeleteInstance deletes an instance
func (m *DockerManager) DeleteInstance(ctx context.Context, instanceID uuid.UUID) error {
	// Get the instance from the database
	instance, err := db.GetInstanceByID(instanceID)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}
	
	// Make sure we have a container ID
	if instance.ContainerID == "" {
		return fmt.Errorf("instance has no container ID")
	}
	
	m.logger.WithFields(logrus.Fields{
		"instance_id":  instance.ID,
		"container_id": instance.ContainerID,
	}).Info("Deleting container")
	
	// Stop the container if it's running
	timeout := 30 * time.Second
	_ = m.client.ContainerStop(ctx, instance.ContainerID, &timeout)
	
	// Remove the container
	if err := m.client.ContainerRemove(ctx, instance.ContainerID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
		m.logger.WithError(err).Error("Failed to remove container")
		return fmt.Errorf("failed to remove container: %w", err)
	}
	
	// Remove host bind mount directory if it exists
	dataDir := m.generateDataDir(fmt.Sprintf("n8n-%s", instance.ID.String()[:8]))
	if err := os.RemoveAll(dataDir); err != nil {
		m.logger.WithError(err).Warn("Failed to remove data directory")
		// Non-fatal error, continue
	}
	
	// Delete DNS record
	subdomain := instance.Host
	dockerDNS := fmt.Sprintf("%s.docker", subdomain)
	
	// Check if DNS record exists before attempting to delete
	records, listErr := m.dnsManager.GetDNSRewrites()
	if listErr != nil {
		m.logger.WithError(listErr).Error("Failed to list DNS records before deletion")
	} else {
		recordExists := false
		var recordIP string
		for _, record := range records {
			if record.Domain == dockerDNS {
				recordExists = true
				recordIP = record.Answer
				m.logger.WithFields(logrus.Fields{
					"domain": record.Domain,
					"ip": record.Answer,
				}).Info("Found DNS record that will be deleted")
				break
			}
		}
		
		if !recordExists {
			m.logger.WithField("dns_record", dockerDNS).Warn("DNS record not found before deletion attempt")
		} else {
			// Log DNS deletion attempt with more details
			m.logger.WithFields(logrus.Fields{
				"instance_id": instance.ID,
				"subdomain": subdomain,
				"dns_record": dockerDNS,
				"ip": recordIP,
			}).Info("Attempting to delete DNS record")
			
			// Try to delete the DNS record
			err := m.dnsManager.DeleteDNSRewrite(dockerDNS)
			if err != nil {
				m.logger.WithFields(logrus.Fields{
					"error": err.Error(),
					"dns_record": dockerDNS,
				}).Warn("Failed to delete DNS record via API, but continuing with instance deletion")
				
				// Add a TODO note about this in the logs
				m.logger.Warn("TODO: Manually clean up DNS record or implement a reliable AdGuard DNS API for deletions")
			} else {
				m.logger.WithField("dns_record", dockerDNS).Info("Successfully deleted DNS record")
			}
		}
	}
	
	// Update instance status
	instance.Status = models.StatusDeleted
	if err := db.UpdateInstance(instance); err != nil {
		m.logger.WithError(err).Warn("Failed to update instance status")
		// We still return success since the container was deleted
	}
	
	m.logger.WithField("instance_id", instance.ID).Info("Container deleted successfully")
	return nil
}

// GetInstanceStats retrieves resource usage stats for an instance
func (m *DockerManager) GetInstanceStats(ctx context.Context, instanceID uuid.UUID) (*models.ResourceUsage, error) {
	// Get the instance from the database
	instance, err := db.GetInstanceByID(instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}
	
	// Make sure we have a container ID
	if instance.ContainerID == "" {
		return nil, fmt.Errorf("instance has no container ID")
	}
	
	m.logger.WithFields(logrus.Fields{
		"instance_id":  instance.ID,
		"container_id": instance.ContainerID,
	}).Debug("Fetching container stats")
	
	// Get container stats
	stats, err := m.client.ContainerStats(ctx, instance.ContainerID, false)
	if err != nil {
		m.logger.WithError(err).Error("Failed to get container stats")
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer stats.Body.Close()
	
	// Parse the stats
	var statsJSON types.StatsJSON
	if err := json.NewDecoder(stats.Body).Decode(&statsJSON); err != nil {
		m.logger.WithError(err).Error("Failed to decode stats")
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
	memoryLimit := statsJSON.MemoryStats.Limit
	memoryPercentage := 0.0
	if memoryLimit > 0 {
		memoryPercentage = (float64(memoryUsage) / float64(memoryLimit)) * 100.0
	}
	
	// Calculate network stats
	var networkIn, networkOut int64
	if networks, ok := statsJSON.Networks["eth0"]; ok {
		networkIn = int64(networks.RxBytes)
		networkOut = int64(networks.TxBytes)
	}
	
	// Create resource usage record
	usage := &models.ResourceUsage{
		InstanceID:      instance.ID,
		Timestamp:       time.Now(),
		CPUUsage:        cpuUsage,
		MemoryUsage:     int64(memoryUsage),
		MemoryLimit:     int64(memoryLimit),
		MemoryPercentage: memoryPercentage,
		DiskUsage:       0, // Implement disk usage calculation if needed
		NetworkIn:       networkIn,
		NetworkOut:      networkOut,
	}
	
	// Save the stats to the database
	if err := db.CreateResourceUsage(usage); err != nil {
		m.logger.WithError(err).Warn("Failed to save resource usage to database")
		// Still return the stats even if saving fails
	}
	
	m.logger.WithFields(logrus.Fields{
		"instance_id": instance.ID,
		"cpu_usage":   fmt.Sprintf("%.2f%%", cpuUsage),
		"memory_usage": fmt.Sprintf("%.2f MB / %.2f MB (%.2f%%)", 
			float64(memoryUsage)/(1024*1024), 
			float64(memoryLimit)/(1024*1024),
			memoryPercentage),
	}).Debug("Container stats collected successfully")
	
	return usage, nil
}

// ListInstances lists all containers managed by LaunchStack
func (m *DockerManager) ListInstances(ctx context.Context) ([]types.Container, error) {
	filters := filters.NewArgs()
	filters.Add("label", "com.launchstack.managed=true")
	
	return m.client.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: filters,
	})
}

// GetInstanceByID retrieves a container by instance ID
func (m *DockerManager) GetInstanceByID(ctx context.Context, instanceID uuid.UUID) (types.Container, error) {
	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("com.launchstack.instance.id=%s", instanceID.String()))
	
	containers, err := m.client.ContainerList(ctx, types.ContainerListOptions{
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
func (m *DockerManager) GetInstancesByUserID(ctx context.Context, userID uuid.UUID) ([]types.Container, error) {
	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("com.launchstack.user.id=%s", userID.String()))
	
	return m.client.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: filters,
	})
} 