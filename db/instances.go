package db

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/launchstack/backend/models"
	"github.com/sirupsen/logrus"
)

// Logger is a package-level logger that can be set by the caller
var Logger *logrus.Logger

func getLogger() *logrus.Logger {
	if Logger == nil {
		Logger = logrus.New()
		Logger.Info("Created default logger for db package")
	}
	return Logger
}

// GetInstancesByUserID retrieves all instances for a user
func GetInstancesByUserID(userID uuid.UUID) ([]models.Instance, error) {
	logger := getLogger()
	logger.WithField("user_id", userID).Info("Fetching instances for user")
	
	var instances []models.Instance
	if err := DB.Where("user_id = ?", userID).Find(&instances).Error; err != nil {
		logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to fetch instances from database")
		return nil, fmt.Errorf("failed to get instances: %w", err)
	}
	
	logger.WithFields(logrus.Fields{
		"user_id": userID,
		"count":   len(instances),
	}).Info("Successfully fetched instances from database")
	
	return instances, nil
}

// GetInstanceByID retrieves an instance by ID
func GetInstanceByID(instanceID uuid.UUID) (*models.Instance, error) {
	logger := getLogger()
	logger.WithField("instance_id", instanceID).Info("Fetching instance by ID")
	
	var instance models.Instance
	if err := DB.Where("id = ?", instanceID).First(&instance).Error; err != nil {
		logger.WithFields(logrus.Fields{
			"instance_id": instanceID,
			"error":       err.Error(),
		}).Error("Failed to fetch instance from database")
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}
	
	logger.WithFields(logrus.Fields{
		"instance_id": instanceID,
		"name":        instance.Name,
		"status":      instance.Status,
	}).Info("Successfully fetched instance from database")
	
	return &instance, nil
}

// CreateInstance creates a new instance
func CreateInstance(instance *models.Instance) error {
	logger := getLogger()
	logger.WithFields(logrus.Fields{
		"instance_id": instance.ID,
		"user_id":     instance.UserID,
		"name":        instance.Name,
	}).Info("Creating new instance in database")
	
	if err := DB.Create(instance).Error; err != nil {
		logger.WithFields(logrus.Fields{
			"instance_id": instance.ID,
			"user_id":     instance.UserID,
			"error":       err.Error(),
		}).Error("Failed to create instance in database")
		return fmt.Errorf("failed to create instance: %w", err)
	}
	
	logger.WithFields(logrus.Fields{
		"instance_id": instance.ID,
		"user_id":     instance.UserID,
		"name":        instance.Name,
		"status":      instance.Status,
	}).Info("Successfully created instance in database")
	
	return nil
}

// UpdateInstance updates an existing instance
func UpdateInstance(instance *models.Instance) error {
	logger := getLogger()
	logger.WithFields(logrus.Fields{
		"instance_id": instance.ID,
		"name":        instance.Name,
		"status":      instance.Status,
	}).Info("Updating instance in database")
	
	if err := DB.Save(instance).Error; err != nil {
		logger.WithFields(logrus.Fields{
			"instance_id": instance.ID,
			"error":       err.Error(),
		}).Error("Failed to update instance in database")
		return fmt.Errorf("failed to update instance: %w", err)
	}
	
	logger.WithFields(logrus.Fields{
		"instance_id": instance.ID,
		"name":        instance.Name,
		"status":      instance.Status,
	}).Info("Successfully updated instance in database")
	
	return nil
}

// DeleteInstance deletes an instance
func DeleteInstance(instanceID uuid.UUID) error {
	logger := getLogger()
	logger.WithField("instance_id", instanceID).Info("Deleting instance from database")
	
	if err := DB.Delete(&models.Instance{}, instanceID).Error; err != nil {
		logger.WithFields(logrus.Fields{
			"instance_id": instanceID,
			"error":       err.Error(),
		}).Error("Failed to delete instance from database")
		return fmt.Errorf("failed to delete instance: %w", err)
	}
	
	logger.WithField("instance_id", instanceID).Info("Successfully deleted instance from database")
	
	return nil
}

// CountInstancesByUserID counts how many instances a user has
func CountInstancesByUserID(userID uuid.UUID) (int64, error) {
	logger := getLogger()
	logger.WithField("user_id", userID).Info("Counting instances for user")
	
	var count int64
	if err := DB.Model(&models.Instance{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to count instances from database")
		return 0, fmt.Errorf("failed to count instances: %w", err)
	}
	
	logger.WithFields(logrus.Fields{
		"user_id": userID,
		"count":   count,
	}).Info("Successfully counted instances from database")
	
	return count, nil
}

// GetRunningInstances retrieves all instances with running status
func GetRunningInstances() ([]models.Instance, error) {
	var instances []models.Instance
	result := DB.Where("status = ?", models.StatusRunning).Find(&instances)
	return instances, result.Error
} 