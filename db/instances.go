package db

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/launchstack/backend/models"
)

// GetInstancesByUserID retrieves all instances for a user
func GetInstancesByUserID(userID uuid.UUID) ([]models.Instance, error) {
	var instances []models.Instance
	result := DB.Where("user_id = ?", userID).Find(&instances)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get instances: %w", result.Error)
	}
	return instances, nil
}

// GetInstanceByID retrieves an instance by ID
func GetInstanceByID(instanceID uuid.UUID) (*models.Instance, error) {
	var instance models.Instance
	result := DB.Where("id = ?", instanceID).First(&instance)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get instance: %w", result.Error)
	}
	return &instance, nil
}

// CreateInstance creates a new instance
func CreateInstance(instance *models.Instance) error {
	result := DB.Create(instance)
	if result.Error != nil {
		return fmt.Errorf("failed to create instance: %w", result.Error)
	}
	return nil
}

// UpdateInstance updates an existing instance
func UpdateInstance(instance *models.Instance) error {
	result := DB.Save(instance)
	if result.Error != nil {
		return fmt.Errorf("failed to update instance: %w", result.Error)
	}
	return nil
}

// DeleteInstance deletes an instance
func DeleteInstance(instanceID uuid.UUID) error {
	result := DB.Delete(&models.Instance{}, "id = ?", instanceID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete instance: %w", result.Error)
	}
	return nil
}

// CountInstancesByUserID counts how many instances a user has
func CountInstancesByUserID(userID uuid.UUID) (int64, error) {
	var count int64
	result := DB.Model(&models.Instance{}).Where("user_id = ?", userID).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count instances: %w", result.Error)
	}
	return count, nil
} 