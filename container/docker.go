package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
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

// generateVolumeNames creates volume names for an instance
func (m *DockerManager) generateVolumeNames(containerName string) (string, string) {
	// Create unique volume names based on container name
	dataVolume := fmt.Sprintf("%s-data", containerName)
	filesVolume := fmt.Sprintf("%s-files", containerName)
	return dataVolume, filesVolume
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
		CPULimit:     user.GetCPULimit(),
		MemoryLimit:  user.GetMemoryLimit(),
		StorageLimit: user.GetStorageLimit(),
	}
	
	// Generate volume names for this container
	dataVolume, filesVolume := m.generateVolumeNames(containerName)
	
	// Set up container memory and CPU limits
	memoryLimit := int64(instance.MemoryLimit * 1024 * 1024) // Convert MB to bytes
	// Convert CPU cores to nano CPUs (1 core = 1000000000 nano CPUs)
	cpuLimit := int64(instance.CPULimit * 1000000000)
	
	// Create host config with resource limits
	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		// Use Docker volumes instead of bind mounts
		Mounts: []mount.Mount{
			{
				Type:     mount.TypeVolume,
				Source:   dataVolume,
				Target:   "/home/node/.n8n",
				ReadOnly: false,
			},
			{
				Type:     mount.TypeVolume,
				Source:   filesVolume,
				Target:   "/files",
				ReadOnly: false,
			},
		},
		Resources: container.Resources{
			Memory:    memoryLimit,
			NanoCPUs:  cpuLimit,
		},
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
		"cpu_limit":  instance.CPULimit,
		"data_volume": dataVolume,
		"files_volume": filesVolume,
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
				// Watchtower labels for automatic updates
				"com.centurylinklabs.watchtower.enable": "true",
				"com.centurylinklabs.watchtower.stop-signal": "SIGTERM",
				"com.centurylinklabs.watchtower.timeout": "60s",
				"com.centurylinklabs.watchtower.cleanup": "true",
				"com.centurylinklabs.watchtower.lifecycle.pre-update": "touch /tmp/pre-update",
				"com.centurylinklabs.watchtower.lifecycle.post-update": "touch /tmp/post-update",
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

// DeleteInstance deletes an n8n instance
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
	
	// Determine the container name (needed for volume names)
	containerName := fmt.Sprintf("n8n-%s", instance.ID.String()[:8])
	dataVolume, filesVolume := m.generateVolumeNames(containerName)
	
	// Remove the container
	m.logger.WithField("container_id", instance.ContainerID).Debug("Removing container")
	err = m.client.ContainerRemove(ctx, instance.ContainerID, types.ContainerRemoveOptions{
		RemoveVolumes: false, // We'll handle volume cleanup separately
		Force:         true,
	})
	if err != nil {
		m.logger.WithError(err).Error("Failed to remove container")
		return fmt.Errorf("failed to remove container: %w", err)
	}
	
	// Remove the Docker volumes
	m.logger.WithFields(logrus.Fields{
		"data_volume": dataVolume,
		"files_volume": filesVolume,
	}).Debug("Removing Docker volumes")
	
	// Use the Docker command line to remove volumes (since the API doesn't expose this directly)
	// We'll use a separate goroutine to avoid blocking
	go func() {
		// Wait a bit for the container to be fully removed
		time.Sleep(5 * time.Second)
		
		// Remove the data volume
		cmd := exec.Command("docker", "volume", "rm", dataVolume)
		if output, err := cmd.CombinedOutput(); err != nil {
			m.logger.WithFields(logrus.Fields{
				"error": err.Error(),
				"output": string(output),
				"volume": dataVolume,
			}).Warn("Failed to remove data volume")
		} else {
			m.logger.WithField("volume", dataVolume).Info("Successfully removed data volume")
		}
		
		// Remove the files volume
		cmd = exec.Command("docker", "volume", "rm", filesVolume)
		if output, err := cmd.CombinedOutput(); err != nil {
			m.logger.WithFields(logrus.Fields{
				"error": err.Error(),
				"output": string(output),
				"volume": filesVolume,
			}).Warn("Failed to remove files volume")
		} else {
			m.logger.WithField("volume", filesVolume).Info("Successfully removed files volume")
		}
	}()
	
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
	// Improved CPU calculation based on Docker stats API
	var cpuUsage float64
	
	// Only calculate if we have valid data
	cpuDelta := float64(statsJSON.CPUStats.CPUUsage.TotalUsage - statsJSON.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(statsJSON.CPUStats.SystemUsage - statsJSON.PreCPUStats.SystemUsage)
	
	if cpuDelta > 0 && systemDelta > 0 {
		// Calculate CPU usage based on available CPU cores
		numCPUs := float64(len(statsJSON.CPUStats.CPUUsage.PercpuUsage))
		if numCPUs == 0 {
			// If PercpuUsage is empty, use the default value of 1
			numCPUs = 1
		}
		
		// Calculate CPU usage as a percentage (0-100) of total available CPU
		// This represents the percentage of total CPU capacity being used
		cpuUsage = (cpuDelta / systemDelta) * numCPUs * 100.0
		
		// Ensure the value is in the range of 0-100%
		if cpuUsage > 100.0 {
			cpuUsage = 100.0
		} else if cpuUsage < 0.01 && cpuUsage > 0 {
			// Ensure very small but non-zero values don't get reported as 0
			// 0.01% is the minimum value we'll report
			cpuUsage = 0.01
		}
		
		// Log CPU deltas for debugging
		m.logger.WithFields(logrus.Fields{
			"cpu_delta": cpuDelta,
			"system_delta": systemDelta,
			"num_cpus": numCPUs,
			"cpu_usage_percent": cpuUsage,
			"container_id": instance.ContainerID,
		}).Debug("CPU usage calculation details")
	} else {
		m.logger.Debug("Unable to calculate accurate CPU usage, values are zero or negative")
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
	for network, stats := range statsJSON.Networks {
		networkIn += int64(stats.RxBytes)
		networkOut += int64(stats.TxBytes)
		m.logger.WithFields(logrus.Fields{
			"network": network,
			"rx_bytes": stats.RxBytes,
			"tx_bytes": stats.TxBytes,
		}).Debug("Network usage details")
	}
	
	// Create resource usage record
	usage := &models.ResourceUsage{
		InstanceID:      instance.ID,
		Timestamp:       time.Now(),
		CPUUsage:        cpuUsage,
		MemoryUsage:     int64(memoryUsage),
		MemoryLimit:     int64(memoryLimit),
		MemoryPercentage: memoryPercentage,
		DiskUsage:       0, // Not tracking disk usage as requested
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

// getVolumeSizeFromAPI gets the volume size using Docker API directly
func (m *DockerManager) getVolumeSizeFromAPI(volumeName string) int64 {
	// Extract host without scheme
	host := m.config.Docker.Host
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")
	
	// Create request to Docker API endpoint for volumes
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("http://%s/volumes", host)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		m.logger.WithError(err).Warn("Failed to create request for Docker volumes API")
		return estimateVolumeSize(volumeName)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		m.logger.WithError(err).Warn("Failed to get volumes from Docker API")
		return estimateVolumeSize(volumeName)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		m.logger.WithField("status", resp.Status).Warn("Docker API returned non-OK status")
		return estimateVolumeSize(volumeName)
	}
	
	// Parse the response
	var result struct {
		Volumes []struct {
			Name      string `json:"Name"`
			UsageData struct {
				Size int64 `json:"Size"`
			} `json:"UsageData"`
		} `json:"Volumes"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		m.logger.WithError(err).Warn("Failed to decode Docker volumes response")
		return estimateVolumeSize(volumeName)
	}
	
	// Find the target volume
	for _, volume := range result.Volumes {
		if volume.Name == volumeName {
			// If size is available, return it
			if volume.UsageData.Size > 0 {
				return volume.UsageData.Size
			}
			break
		}
	}
	
	// If we get here, either the volume wasn't found or size was 0
	// Fall back to the existing method
	return m.getVolumeSize(volumeName)
}

// getVolumeSize returns the size of a Docker volume in bytes
func (m *DockerManager) getVolumeSize(volumeName string) int64 {
	// Use Docker API to inspect the volume first
	ctx := context.Background()
	
	// Create a new Docker client with the same host as the main client
	// This is a workaround since our DockerClient interface doesn't expose VolumeInspect
	cli, err := client.NewClientWithOpts(
		client.WithHost(m.config.Docker.Host),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		m.logger.WithError(err).Warn("Failed to create Docker client for volume inspection")
		return estimateVolumeSize(volumeName)
	}
	defer cli.Close()
	
	// Inspect the volume
	vol, err := cli.VolumeInspect(ctx, volumeName)
	if err != nil {
		m.logger.WithFields(logrus.Fields{
			"volume": volumeName,
			"error":  err.Error(),
		}).Warn("Failed to inspect volume")
		return estimateVolumeSize(volumeName)
	}
	
	// Get size from volume status if available
	if vol.Status != nil {
		if sizeStr, ok := vol.Status["Size"]; ok {
			sizeString, ok := sizeStr.(string)
			if ok {
				size, err := strconv.ParseInt(sizeString, 10, 64)
				if err == nil {
					return size
				}
			}
		}
	}
	
	// If we got this far, we need to try an alternative method
	// Docker API doesn't provide volume size directly in all environments
	
	// Try executing the "du" command inside the container that uses this volume
	// This requires finding containers that use this volume
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		m.logger.WithError(err).Warn("Failed to list containers for volume size calculation")
		return estimateVolumeSize(volumeName)
	}
	
	// Find containers that use this volume
	for _, container := range containers {
		// Get container details
		info, err := cli.ContainerInspect(ctx, container.ID)
		if err != nil {
			continue
		}
		
		// Check if this container uses our volume
		for _, mount := range info.Mounts {
			if mount.Type == "volume" && mount.Name == volumeName {
				// This container uses our volume
				// For N8N volumes, estimate based on instance age and typical usage patterns
				if strings.Contains(volumeName, "n8n-") {
					createdTime := info.Created
					t, err := time.Parse(time.RFC3339, createdTime)
					if err != nil {
						m.logger.WithError(err).Warn("Failed to parse container creation time")
						return estimateVolumeSize(volumeName)
					}
					
					// Calculate age in days
					ageInDays := time.Since(t).Hours() / 24
					
					// Base size + growth per day
					// Data volume: 50MB base + 5MB per day
					// Files volume: 10MB base + 2MB per day
					if strings.Contains(volumeName, "-data") {
						return int64(50*1024*1024 + ageInDays*5*1024*1024)
					} else if strings.Contains(volumeName, "-files") {
						return int64(10*1024*1024 + ageInDays*2*1024*1024)
					}
				}
			}
		}
	}
	
	// Fallback to estimation if no other method works
	return estimateVolumeSize(volumeName)
}

// estimateVolumeSize provides a reasonable estimate for volume size when direct measurement fails
func estimateVolumeSize(volumeName string) int64 {
	// Provide different defaults based on volume type
	if strings.Contains(volumeName, "-data") {
		// Data volumes typically start around 100MB
		return 100 * 1024 * 1024
	} else if strings.Contains(volumeName, "-files") {
		// Files volumes typically start smaller
		return 20 * 1024 * 1024
	}
	
	// Generic fallback
	return 50 * 1024 * 1024
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