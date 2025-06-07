package container

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/launchstack/backend/models"
	"github.com/sirupsen/logrus"
)

// Manager handles container operations
type Manager struct {
	logger *logrus.Logger
}

// NewManager creates a new container manager
func NewManager(logger *logrus.Logger) *Manager {
	return &Manager{
		logger: logger,
	}
}

// CreateInstance creates a new instance (mock implementation)
func (m *Manager) CreateInstance(ctx context.Context, user models.User, instanceReq models.Instance) (*models.Instance, error) {
	m.logger.Info("Mock: Creating instance")
	instance := &models.Instance{
		ID:          uuid.New(),
		UserID:      user.ID,
		Name:        instanceReq.Name,
		Description: instanceReq.Description,
		Status:      models.StatusRunning,
		Host:        fmt.Sprintf("%s-%s", instanceReq.Name, user.ID.String()[:8]),
		Port:        5000,
		URL:         fmt.Sprintf("https://%s-%s.launchstack.io", instanceReq.Name, user.ID.String()[:8]),
		CPULimit:    1.0,
		MemoryLimit: 1024,
		StorageLimit: 10,
	}
	return instance, nil
}

// DeleteInstance deletes an instance (mock implementation)
func (m *Manager) DeleteInstance(ctx context.Context, instanceID uuid.UUID) error {
	m.logger.Infof("Mock: Deleting instance %s", instanceID.String())
	return nil
}

// StartInstance starts an instance (mock implementation)
func (m *Manager) StartInstance(ctx context.Context, instanceID uuid.UUID) error {
	m.logger.Infof("Mock: Starting instance %s", instanceID.String())
	return nil
}

// StopInstance stops an instance (mock implementation)
func (m *Manager) StopInstance(ctx context.Context, instanceID uuid.UUID) error {
	m.logger.Infof("Mock: Stopping instance %s", instanceID.String())
	return nil
} 