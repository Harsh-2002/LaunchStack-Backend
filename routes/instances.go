package routes

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/launchstack/backend/container"
	"github.com/launchstack/backend/db"
	"github.com/launchstack/backend/middleware"
	"github.com/launchstack/backend/models"
)

// InstanceRequest is the request body for creating/updating an instance
type InstanceRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// GetInstances returns all instances for the current user
func GetInstances(containerManager *container.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, err := middleware.GetUserIDFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// Get instances from database
		instances, err := db.GetInstancesByUserID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get instances"})
			return
		}

		// Convert to response format
		response := make([]map[string]interface{}, len(instances))
		for i, instance := range instances {
			response[i] = instance.ToPublicResponse()
		}

		c.JSON(http.StatusOK, response)
	}
}

// CreateInstance creates a new instance
func CreateInstance(containerManager *container.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		user, err := middleware.GetUserFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// Parse request body
		var req InstanceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Check if user has reached their instance limit
		count, err := db.CountInstancesByUserID(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check instance count"})
			return
		}

		if int(count) >= user.GetInstancesLimit() {
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
		instance, err := containerManager.CreateInstance(context.Background(), user, instanceReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create instance: " + err.Error()})
			return
		}

		// Save instance to database
		if err := db.CreateInstance(instance); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save instance"})
			return
		}

		c.JSON(http.StatusCreated, instance.ToPublicResponse())
	}
}

// GetInstance returns a specific instance
func GetInstance(containerManager *container.Manager) gin.HandlerFunc {
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

		c.JSON(http.StatusOK, instance.ToPublicResponse())
	}
}

// UpdateInstance updates an existing instance
func UpdateInstance(containerManager *container.Manager) gin.HandlerFunc {
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
func DeleteInstance(containerManager *container.Manager) gin.HandlerFunc {
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
func StartInstance(containerManager *container.Manager) gin.HandlerFunc {
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
func StopInstance(containerManager *container.Manager) gin.HandlerFunc {
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
func RestartInstance(containerManager *container.Manager) gin.HandlerFunc {
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