package db

import (
	"github.com/google/uuid"
	"github.com/launchstack/backend/models"
)

// CreateResourceUsage saves a resource usage record to the database
func CreateResourceUsage(usage *models.ResourceUsage) error {
	result := DB.Create(usage)
	return result.Error
}

// GetResourceUsageByInstanceID retrieves resource usage records for an instance
func GetResourceUsageByInstanceID(instanceID uuid.UUID, limit int) ([]models.ResourceUsage, error) {
	var usages []models.ResourceUsage
	
	// Set a default limit if not specified
	if limit <= 0 {
		limit = 10
	}
	
	result := DB.Where("instance_id = ?", instanceID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&usages)
	
	return usages, result.Error
}

// GetLatestResourceUsage retrieves the most recent resource usage record for an instance
func GetLatestResourceUsage(instanceID uuid.UUID) (*models.ResourceUsage, error) {
	var usage models.ResourceUsage
	
	result := DB.Where("instance_id = ?", instanceID).
		Order("timestamp DESC").
		First(&usage)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return &usage, nil
}

// PruneResourceUsage deletes old resource usage records to prevent excessive database growth
func PruneResourceUsage(maxRecordsPerInstance int) error {
	// Get all instance IDs
	var instanceIDs []uuid.UUID
	if err := DB.Model(&models.ResourceUsage{}).
		Distinct("instance_id").
		Pluck("instance_id", &instanceIDs).Error; err != nil {
		return err
	}
	
	// For each instance, keep only the most recent records
	for _, instanceID := range instanceIDs {
		// Find the IDs to keep
		var idsToKeep []uuid.UUID
		if err := DB.Model(&models.ResourceUsage{}).
			Where("instance_id = ?", instanceID).
			Order("timestamp DESC").
			Limit(maxRecordsPerInstance).
			Pluck("id", &idsToKeep).Error; err != nil {
			continue // Skip this instance if there's an error
		}
		
		// Delete records not in the keep list
		if len(idsToKeep) > 0 {
			DB.Where("instance_id = ? AND id NOT IN ?", instanceID, idsToKeep).
				Delete(&models.ResourceUsage{})
		}
	}
	
	return nil
} 