package container

import (
	"context"

	"github.com/google/uuid"
	"github.com/launchstack/backend/models"
)

// Manager is the interface for container operations
type Manager interface {
	// CreateInstance creates a new instance
	CreateInstance(ctx context.Context, user models.User, instanceReq models.Instance) (*models.Instance, error)
	
	// DeleteInstance deletes an instance
	DeleteInstance(ctx context.Context, instanceID uuid.UUID) error
	
	// StartInstance starts an instance
	StartInstance(ctx context.Context, instanceID uuid.UUID) error
	
	// StopInstance stops an instance
	StopInstance(ctx context.Context, instanceID uuid.UUID) error
	
	// GetInstanceStats retrieves resource usage stats for an instance
	GetInstanceStats(ctx context.Context, instanceID uuid.UUID) (*models.ResourceUsage, error)
} 