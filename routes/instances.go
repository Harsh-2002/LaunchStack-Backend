package routes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/launchstack/backend/container"
	"github.com/launchstack/backend/db"
	"github.com/launchstack/backend/middleware"
	"github.com/launchstack/backend/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InstanceRequest is the request body for creating/updating an instance
type InstanceRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// GetInstances returns all instances for the current user
func GetInstances(containerManager container.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get logger from context
		logger := c.MustGet("logger").(*logrus.Logger)
		logger.Info("Received request to get all instances")
		
		// Get user ID from context
		userID, err := middleware.GetUserIDFromContext(c)
		if err != nil {
			logger.WithError(err).Error("Failed to get user ID from context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		logger.WithField("user_id", userID).Info("Processing get instances request for user")

		// Get instances from database
		logger.Info("Fetching instances from database")
		instances, err := db.GetInstancesByUserID(userID)
		if err != nil {
			logger.WithError(err).Error("Failed to get instances from database")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get instances"})
			return
		}
		logger.WithField("instance_count", len(instances)).Info("Successfully retrieved instances")

		// Convert to response format
		logger.Info("Preparing response")
		response := make([]map[string]interface{}, len(instances))
		for i, instance := range instances {
			response[i] = instance.ToPublicResponse()
			logger.WithFields(logrus.Fields{
				"instance_id":   instance.ID,
				"instance_name": instance.Name,
				"status":        instance.Status,
				"url":           instance.URL,
			}).Debug("Added instance to response")
		}

		logger.WithField("response_count", len(response)).Info("Returning instances to client")
		c.JSON(http.StatusOK, response)
	}
}

// CreateInstance creates a new instance
func CreateInstance(containerManager container.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := c.MustGet("logger").(*logrus.Logger)
		logger.Info("Received request to create a new instance")
		
		// Get user from context
		user, err := middleware.GetUserFromContext(c)
		if err != nil {
			logger.WithError(err).Error("Failed to get user from context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		
		// Debug the user ID
		logger.WithFields(logrus.Fields{
			"user_id":    user.ID.String(),
			"user_email": user.Email,
			"plan":       user.Plan,
		}).Info("Processing instance creation for user")
		
		// Parse request body
		var req InstanceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.WithError(err).Error("Invalid request body")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		logger.WithFields(logrus.Fields{
			"instance_name": req.Name,
			"description":   req.Description,
		}).Info("Received instance creation parameters")

		// Check if user has reached their instance limit
		count, err := db.CountInstancesByUserID(user.ID)
		if err != nil {
			logger.WithError(err).Error("Failed to check instance count")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check instance count"})
			return
		}

		logger.WithFields(logrus.Fields{
			"current_count": count,
			"limit":         user.GetInstancesLimit(),
		}).Info("Checking instance limits")
		
		if int(count) >= user.GetInstancesLimit() {
			logger.WithFields(logrus.Fields{
				"current_count": count,
				"limit":         user.GetInstancesLimit(),
			}).Warn("Instance limit reached")
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Instance limit reached",
				"limit": user.GetInstancesLimit(),
			})
			return
		}

		// Create instance request object
		instanceReq := models.Instance{
			Name:        req.Name,
			Description: req.Description,
		}

		// Create the instance
		logger.Info("Calling container manager to create instance")
		instance, err := containerManager.CreateInstance(context.Background(), user, instanceReq)
		if err != nil {
			logger.WithError(err).Error("Failed to create instance")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create instance: " + err.Error()})
			return
		}
		logger.WithFields(logrus.Fields{
			"instance_id":   instance.ID,
			"instance_name": instance.Name,
			"status":        instance.Status,
			"url":           instance.URL,
		}).Info("Instance created successfully by container manager")

		// Save instance to database
		logger.Info("Saving instance to database")
		if err := db.CreateInstance(instance); err != nil {
			logger.WithError(err).Error("Failed to save instance to database")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save instance"})
			return
		}
		logger.WithField("instance_id", instance.ID).Info("Instance saved to database")

		c.JSON(http.StatusCreated, instance.ToPublicResponse())
		logger.WithField("instance_id", instance.ID).Info("Instance creation completed successfully")
	}
}

// GetInstance returns a specific instance
func GetInstance(containerManager container.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get logger from context
		logger := c.MustGet("logger").(*logrus.Logger)
		
		// Parse instance ID from URL
		instanceIDStr := c.Param("id")
		logger.WithField("instance_id", instanceIDStr).Info("Received request to get specific instance")
		
		// Get user ID from context
		userID, err := middleware.GetUserIDFromContext(c)
		if err != nil {
			logger.WithError(err).Error("Failed to get user ID from context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		logger.WithField("user_id", userID).Info("Processing get instance request for user")

		instanceID, err := uuid.Parse(instanceIDStr)
		if err != nil {
			logger.WithError(err).Error("Invalid instance ID format")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid instance ID"})
			return
		}

		// Get instance from database
		logger.WithField("instance_id", instanceID).Info("Fetching instance from database")
		instance, err := db.GetInstanceByID(instanceID)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"instance_id": instanceID,
				"error":       err.Error(),
			}).Error("Failed to fetch instance from database")
			c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
			return
		}
		logger.WithFields(logrus.Fields{
			"instance_id":   instance.ID,
			"instance_name": instance.Name,
			"status":        instance.Status,
		}).Info("Successfully retrieved instance")

		// Check if the instance belongs to the user
		if instance.UserID != userID {
			logger.WithFields(logrus.Fields{
				"instance_user_id": instance.UserID,
				"request_user_id":  userID,
			}).Warn("User attempted to access instance they don't own")
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		logger.WithField("instance_id", instance.ID).Info("Returning instance details to client")
		c.JSON(http.StatusOK, instance.ToPublicResponse())
	}
}

// UpdateInstance updates an existing instance
func UpdateInstance(containerManager container.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, err := middleware.GetUserIDFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// Parse instance ID from URL
		instanceIDStr := c.Param("id")
		instanceID, err := uuid.Parse(instanceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid instance ID"})
			return
		}

		// Get instance from database
		instance, err := db.GetInstanceByID(instanceID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
			return
		}

		// Check if the instance belongs to the user
		if instance.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		// Parse request body
		var req InstanceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Update instance properties
		instance.Name = req.Name
		instance.Description = req.Description

		// Save changes to database
		if err := db.UpdateInstance(instance); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update instance"})
			return
		}

		c.JSON(http.StatusOK, instance.ToPublicResponse())
	}
}

// DeleteInstance deletes an instance
func DeleteInstance(containerManager container.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, err := middleware.GetUserIDFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// Parse instance ID from URL
		instanceIDStr := c.Param("id")
		instanceID, err := uuid.Parse(instanceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid instance ID"})
			return
		}

		// Get instance from database
		instance, err := db.GetInstanceByID(instanceID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
			return
		}

		// Check if the instance belongs to the user
		if instance.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		// Delete the container
		if err := containerManager.DeleteInstance(context.Background(), instanceID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete instance container"})
			return
		}

		// Delete from database
		if err := db.DeleteInstance(instanceID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete instance from database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Instance deleted successfully"})
	}
}

// StartInstance starts a stopped instance
func StartInstance(containerManager container.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, err := middleware.GetUserIDFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// Parse instance ID from URL
		instanceIDStr := c.Param("id")
		instanceID, err := uuid.Parse(instanceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid instance ID"})
			return
		}

		// Get instance from database
		instance, err := db.GetInstanceByID(instanceID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
			return
		}

		// Check if the instance belongs to the user
		if instance.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		// Check if the instance is already running
		if instance.Status == models.StatusRunning {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Instance is already running"})
			return
		}

		// Start the instance
		if err := containerManager.StartInstance(context.Background(), instanceID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start instance"})
			return
		}

		// Update instance status
		instance.Status = models.StatusRunning
		if err := db.UpdateInstance(instance); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update instance status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Instance started successfully"})
	}
}

// StopInstance stops a running instance
func StopInstance(containerManager container.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, err := middleware.GetUserIDFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// Parse instance ID from URL
		instanceIDStr := c.Param("id")
		instanceID, err := uuid.Parse(instanceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid instance ID"})
			return
		}

		// Get instance from database
		instance, err := db.GetInstanceByID(instanceID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
			return
		}

		// Check if the instance belongs to the user
		if instance.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		// Check if the instance is already stopped
		if instance.Status == models.StatusStopped {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Instance is already stopped"})
			return
		}

		// Stop the instance
		if err := containerManager.StopInstance(context.Background(), instanceID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop instance"})
			return
		}

		// Update instance status
		instance.Status = models.StatusStopped
		if err := db.UpdateInstance(instance); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update instance status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Instance stopped successfully"})
	}
}

// RestartInstance restarts a running instance
func RestartInstance(containerManager container.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, err := middleware.GetUserIDFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// Parse instance ID from URL
		instanceIDStr := c.Param("id")
		instanceID, err := uuid.Parse(instanceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid instance ID"})
			return
		}

		// Get instance from database
		instance, err := db.GetInstanceByID(instanceID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
			return
		}

		// Check if the instance belongs to the user
		if instance.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		// Stop the instance
		if err := containerManager.StopInstance(context.Background(), instanceID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop instance"})
			return
		}

		// Start the instance
		if err := containerManager.StartInstance(context.Background(), instanceID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start instance"})
			return
		}

		// Update instance status
		instance.Status = models.StatusRunning
		if err := db.UpdateInstance(instance); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update instance status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Instance restarted successfully"})
	}
}

// GetInstanceStats returns resource usage stats for an instance
func GetInstanceStats(containerManager container.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get instance ID from path
		instanceID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid instance ID"})
			return
		}
		
		// Get the user ID from context
		userID, err := middleware.GetUserIDFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		
		// Get the instance from database
		instance, err := db.GetInstanceByID(instanceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching instance"})
			return
		}
		
		// Check if the instance belongs to the user
		if instance.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this instance"})
			return
		}
		
		// Get instance stats
		stats, err := containerManager.GetInstanceStats(context.Background(), instanceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error getting instance stats: %v", err)})
			return
		}
		
		// Return the stats
		c.JSON(http.StatusOK, stats.FormatStats())
	}
}

// GetInstanceHistoricalStats returns historical resource usage for an instance
func GetInstanceHistoricalStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get instance ID from path
		instanceID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid instance ID"})
			return
		}
		
		// Get the user ID from context
		userID, err := middleware.GetUserIDFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		
		// Get the instance from database
		instance, err := db.GetInstanceByID(instanceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching instance"})
			return
		}
		
		// Check if the instance belongs to the user
		if instance.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this instance"})
			return
		}
		
		// Parse query parameters - match frontend expected format
		periodStr := c.DefaultQuery("period", "1h")
		
		// Convert period string to duration
		var period time.Duration
		switch periodStr {
		case "10m":
			period = 10 * time.Minute
		case "1h":
			period = time.Hour
		case "6h":
			period = 6 * time.Hour
		case "24h":
			period = 24 * time.Hour
		default:
			period = time.Hour
		}
		
		// For all periods, use the detailed historical data
		// but format it according to frontend expectations
		metrics, fetchErr := db.GetResourceUsageHistorical(instanceID, period, "auto")
		if fetchErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error fetching metrics: %v", fetchErr)})
			return
		}
		
		// Convert to frontend expected format (plain array of data points)
		dataPoints := make([]map[string]interface{}, 0, len(metrics))
		for _, point := range metrics {
			// Convert the time-bucketed data to match expected frontend format
			dataPoint := map[string]interface{}{
				"timestamp":         point["timestamp"],
				"cpu_usage":         point["cpu_avg"],          // Use average CPU as cpu_usage
				"memory_usage":      point["memory_avg"],       // Use average memory as memory_usage
				"memory_limit":      instance.MemoryLimit,      // Use instance memory limit
				"memory_percentage": point["memory_percentage"], // Use calculated percentage
				"network_in":        point["network_in"],
				"network_out":       point["network_out"],
			}
			dataPoints = append(dataPoints, dataPoint)
		}
		
		// Limit to 100 data points as expected by frontend
		if len(dataPoints) > 100 {
			dataPoints = dataPoints[:100]
		}
		
		// Return just the data points array as expected by frontend
		c.JSON(http.StatusOK, dataPoints)
	}
} 