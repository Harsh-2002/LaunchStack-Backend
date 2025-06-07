package container

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/launchstack/backend/config"
	"github.com/launchstack/backend/db"
	"github.com/launchstack/backend/models"
	"github.com/sirupsen/logrus"
)

// MockManager handles container operations in mock mode
type MockManager struct {
	logger *logrus.Logger
	subnet string
	baseIP net.IP
	domain string
	config *config.Config
	// Track allocated IPs
	allocatedIPs map[string]bool
}

// NewMockManager creates a new mock container manager
func NewMockManager(logger *logrus.Logger, cfg *config.Config) Manager {
	if logger == nil {
		logger = logrus.New()
		logger.Info("Created default logger for container manager")
	}
	logger.Info("Initializing container manager in mock mode")
	
	// Get subnet from config
	subnet := cfg.Docker.NetworkSubnet
	if subnet == "" {
		subnet = "10.1.2.0/24" // Default subnet
		logger.Warnf("DOCKER_NETWORK_SUBNET not set, using default: %s", subnet)
	}
	
	// Parse the subnet
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		logger.Warnf("Invalid subnet %s, using default 10.1.2.0/24", subnet)
		subnet = "10.1.2.0/24"
		_, ipNet, _ = net.ParseCIDR(subnet)
	}
	
	baseIP := ipNet.IP
	
	// Get domain from config
	domain := cfg.Server.Domain
	if domain == "" {
		domain = "srvr.site" // Default domain
		logger.Warnf("DOMAIN not set, using default: %s", domain)
	}
	
	logger.Infof("Container manager initialized with subnet %s and domain %s", subnet, domain)
	
	return &MockManager{
		logger:       logger,
		subnet:       subnet,
		baseIP:       baseIP,
		domain:       domain,
		config:       cfg,
		allocatedIPs: make(map[string]bool),
	}
}

// allocateIP allocates a unique IP address from the subnet
func (m *MockManager) allocateIP() (string, error) {
	// Parse the subnet
	_, ipNet, err := net.ParseCIDR(m.subnet)
	if err != nil {
		return "", fmt.Errorf("invalid subnet %s: %w", m.subnet, err)
	}
	
	// Get the first usable IP in the subnet (skip network address)
	ip := make(net.IP, len(ipNet.IP))
	copy(ip, ipNet.IP)
	
	// Start from the 10th IP in the subnet to avoid conflicts with gateway, etc.
	for i := 0; i < 10; i++ {
		incrementIP(ip)
	}
	
	// Try up to 240 IPs in the subnet
	for i := 0; i < 240; i++ {
		ipStr := ip.String()
		
		if !m.allocatedIPs[ipStr] {
			m.allocatedIPs[ipStr] = true
			return ipStr, nil
		}
		
		incrementIP(ip)
		
		// Check if we've gone outside the subnet
		if !ipNet.Contains(ip) {
			return "", fmt.Errorf("no available IPs in subnet %s", m.subnet)
		}
	}
	
	return "", fmt.Errorf("no available IPs in subnet %s", m.subnet)
}

// incrementIP increments an IP address by 1
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// Using shared implementation from shared.go

// CreateInstance creates a new instance (mock implementation)
func (m *MockManager) CreateInstance(ctx context.Context, user models.User, instanceReq models.Instance) (*models.Instance, error) {
	m.logger.WithFields(logrus.Fields{
		"user_id":       user.ID.String(),
		"user_email":    user.Email,
		"instance_name": instanceReq.Name,
		"plan":          user.Plan,
	}).Info("Creating new instance")
	
	// Log user plan limits
	m.logger.WithFields(logrus.Fields{
		"cpu_limit":     user.GetCPULimit(),
		"memory_limit":  user.GetMemoryLimit(),
		"storage_limit": user.GetStorageLimit(),
		"max_instances": user.GetInstancesLimit(),
	}).Info("User resource limits")
	
	// Generate unique container ID and container name
	instanceID := uuid.New()
	containerName := GenerateContainerName(user.ID, instanceReq.Name)
	
	// Generate a unique, easy-to-remember subdomain
	subdomain := GenerateEasySubdomain(containerName)
	
	// Create unique URLs
	url := fmt.Sprintf("https://%s.%s", subdomain, m.domain)
	
	// Allocate a unique IP
	ip, err := m.allocateIP()
	if err != nil {
		m.logger.WithError(err).Error("Failed to allocate IP address")
		return nil, err
	}
	
	// Get the n8n container port from config
	n8nPort := m.config.Docker.N8NContainerPort
	if n8nPort == 0 {
		n8nPort = 5678 // Default n8n port
	}
	
	m.logger.WithFields(logrus.Fields{
		"instance_id":    instanceID,
		"container_name": containerName,
		"subdomain":      subdomain,
		"url":            url,
		"ip":             ip,
		"n8n_port":       n8nPort,
	}).Info("Generated instance identifiers")
	
	// Use container name for both container ID and volume directory
	// This makes it easier to identify which volumes belong to which container
	dataDir := fmt.Sprintf("/SSD/LaunchStack/N8N/%s/data", containerName)
	filesDir := fmt.Sprintf("/SSD/LaunchStack/N8N/%s/files", containerName)
	
	dockerCmd := fmt.Sprintf(
		"docker run -d "+
			"--name %s "+
			"--restart always "+
			"--user root "+
			"-e N8N_BASIC_AUTH_ACTIVE=true "+
			"-e N8N_HOST=%s.%s "+
			"-e N8N_PROTOCOL=https "+
			"-e NODE_ENV=production "+
			"-e WEBHOOK_URL=https://%s.%s/ "+
			"-v %s:/home/node/.n8n "+
			"-v %s:/files "+
			"--label com.centurylinklabs.watchtower.enable=true "+
			"--network n8n "+
			"--ip %s "+
			"n8nio/n8n:latest",
		containerName,
		subdomain, m.domain,
		subdomain, m.domain,
		dataDir,
		filesDir,
		ip,
	)
	
	m.logger.WithFields(logrus.Fields{
		"docker_cmd": dockerCmd,
	}).Info("Docker command that would be executed")
	
	// In a real implementation, we would now:
	m.logger.Info("MOCK: Would create data directories")
	mockCmd := fmt.Sprintf("mkdir -p %s %s", dataDir, filesDir)
	m.logger.WithField("cmd", mockCmd).Info("Would execute")
	time.Sleep(100 * time.Millisecond)
	
	m.logger.Info("MOCK: Would pull Docker image")
	mockCmd = "docker pull n8nio/n8n:latest"
	m.logger.WithField("cmd", mockCmd).Info("Would execute")
	time.Sleep(100 * time.Millisecond)
	
	m.logger.Info("MOCK: Would create and start container")
	m.logger.WithField("cmd", dockerCmd).Info("Would execute")
	time.Sleep(100 * time.Millisecond)
	
	m.logger.Info("MOCK: Would configure reverse proxy (Caddy)")
	mockCmd = fmt.Sprintf("echo '%s.%s { reverse_proxy http://%s:%d }' >> /etc/caddy/Caddyfile", 
		subdomain, m.domain, ip, n8nPort)
	m.logger.WithField("cmd", mockCmd).Info("Would execute")
	time.Sleep(100 * time.Millisecond)
	
	m.logger.Info("MOCK: Would reload Caddy")
	mockCmd = "systemctl reload caddy"
	m.logger.WithField("cmd", mockCmd).Info("Would execute")
	time.Sleep(100 * time.Millisecond)
	
	// Create the instance object
	instance := &models.Instance{
		ID:           instanceID,
		UserID:       user.ID,
		Name:         instanceReq.Name,
		Description:  instanceReq.Description,
		Status:       models.StatusRunning,
		Host:         subdomain,
		Port:         n8nPort,
		URL:          url,
		CPULimit:     user.GetCPULimit(),
		MemoryLimit:  user.GetMemoryLimit(),
		StorageLimit: user.GetStorageLimit(),
		ContainerID:  containerName, // Use container name as the ID for consistency
		IPAddress:    ip,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	m.logger.WithFields(logrus.Fields{
		"instance_id": instance.ID,
		"status":      instance.Status,
		"url":         instance.URL,
		"ip":          instance.IPAddress,
	}).Info("Instance created successfully")
	
	return instance, nil
}

// DeleteInstance deletes an instance (mock implementation)
func (m *MockManager) DeleteInstance(ctx context.Context, instanceID uuid.UUID) error {
	m.logger.WithFields(logrus.Fields{
		"instance_id": instanceID,
	}).Info("Deleting instance")
	
	// In a real implementation, we would:
	m.logger.Info("MOCK: Would get container details from database")
	time.Sleep(100 * time.Millisecond)
	
	// Use a mock container name based on the instance ID
	containerName := fmt.Sprintf("mock-container-%s", instanceID.String()[:8])
	
	// Generate the Docker commands that would be executed
	dockerStopCmd := fmt.Sprintf("docker stop %s", containerName)
	dockerRmCmd := fmt.Sprintf("docker rm %s", containerName)
	
	m.logger.WithField("cmd", dockerStopCmd).Info("Would execute")
	time.Sleep(100 * time.Millisecond)
	
	m.logger.WithField("cmd", dockerRmCmd).Info("Would execute")
	time.Sleep(100 * time.Millisecond)
	
	// Use container name for volume mount path
	dataDir := fmt.Sprintf("/SSD/LaunchStack/N8N/%s", containerName)
	rmDataCmd := fmt.Sprintf("rm -rf %s", dataDir)
	
	m.logger.WithField("cmd", rmDataCmd).Info("Would execute")
	time.Sleep(100 * time.Millisecond)
	
	m.logger.Info("MOCK: Would update reverse proxy configuration")
	time.Sleep(100 * time.Millisecond)
	
	m.logger.WithFields(logrus.Fields{
		"instance_id": instanceID,
	}).Info("Instance deleted successfully")
	
	return nil
}

// StartInstance starts an instance (mock implementation)
func (m *MockManager) StartInstance(ctx context.Context, instanceID uuid.UUID) error {
	m.logger.WithFields(logrus.Fields{
		"instance_id": instanceID,
	}).Info("Starting instance")
	
	containerName := fmt.Sprintf("mock-container-%s", instanceID.String()[:8])
	dockerStartCmd := fmt.Sprintf("docker start %s", containerName)
	
	m.logger.WithField("cmd", dockerStartCmd).Info("Would execute")
	time.Sleep(100 * time.Millisecond)
	
	m.logger.WithFields(logrus.Fields{
		"instance_id": instanceID,
	}).Info("Instance started successfully")
	
	return nil
}

// StopInstance stops an instance (mock implementation)
func (m *MockManager) StopInstance(ctx context.Context, instanceID uuid.UUID) error {
	m.logger.WithFields(logrus.Fields{
		"instance_id": instanceID,
	}).Info("Mock: Stopping instance")
	
	// Get the instance from the database
	instance, err := db.GetInstanceByID(instanceID)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}
	
	// Update instance status
	instance.Status = models.StatusStopped
	if err := db.UpdateInstance(instance); err != nil {
		m.logger.WithError(err).Warn("Failed to update instance status")
	}
	
	return nil
}

// GetInstanceStats retrieves resource usage stats for an instance (mock implementation)
func (m *MockManager) GetInstanceStats(ctx context.Context, instanceID uuid.UUID) (*models.ResourceUsage, error) {
	m.logger.WithFields(logrus.Fields{
		"instance_id": instanceID,
	}).Info("Mock: Getting instance stats")
	
	// Get the instance from the database
	instance, err := db.GetInstanceByID(instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}
	
	// Generate mock stats
	usage := &models.ResourceUsage{
		InstanceID:       instance.ID,
		Timestamp:        time.Now(),
		CPUUsage:         randomFloat(5, 15),              // Random value between 5-15%
		MemoryUsage:      int64(randomInt(50, 200) * 1024 * 1024), // Random value between 50-200 MB
		MemoryLimit:      int64(instance.MemoryLimit * 1024 * 1024),
		MemoryPercentage: randomFloat(10, 40),            // Random value between 10-40%
		DiskUsage:        int64(randomInt(10, 100) * 1024 * 1024), // Random value between 10-100 MB
		NetworkIn:        int64(randomInt(1000, 10000)),  // Random network traffic
		NetworkOut:       int64(randomInt(1000, 10000)),  // Random network traffic
	}
	
	// Save the stats to the database
	if err := db.CreateResourceUsage(usage); err != nil {
		m.logger.WithError(err).Warn("Failed to save resource usage to database")
		// Still return the stats even if saving fails
	}
	
	return usage, nil
}

// Helper functions for mock data generation
func randomFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
} 