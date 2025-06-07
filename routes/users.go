package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/launchstack/backend/db"
	"github.com/launchstack/backend/middleware"
)

// UserUpdateRequest represents the request to update a user
type UserUpdateRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// GetCurrentUser returns the current authenticated user
func GetCurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := middleware.GetUserFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// Count instances
		instanceCount, err := db.CountInstancesByUserID(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count instances"})
			return
		}

		// Create response with user info and resource allocation
		response := user.ToPublicResponse()
		response["instances"] = map[string]interface{}{
			"current": instanceCount,
			"limit":   user.GetInstancesLimit(),
		}
		response["resource_limits"] = user.GetPlanResourceLimits()

		c.JSON(http.StatusOK, response)
	}
}

// UpdateCurrentUser updates the current authenticated user
func UpdateCurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := middleware.GetUserFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		var req UserUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Update user fields
		if req.FirstName != "" {
			user.FirstName = req.FirstName
		}
		if req.LastName != "" {
			user.LastName = req.LastName
		}

		// Save changes to database
		if err := db.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		c.JSON(http.StatusOK, user.ToPublicResponse())
	}
}

// GetUsageStats returns usage statistics for all user instances
func GetUsageStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := middleware.GetUserIDFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// Get all instances for the user
		instances, err := db.GetInstancesByUserID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get instances"})
			return
		}

		// Placeholder for actual usage stats
		// In a real implementation, you would query resource usage from the database
		usageStats := make(map[string]interface{})
		for _, instance := range instances {
			usageStats[instance.ID.String()] = map[string]interface{}{
				"cpu":     0.0,
				"memory":  0,
				"storage": 0,
				"status":  instance.Status,
			}
		}

		c.JSON(http.StatusOK, usageStats)
	}
}

// GetInstanceUsage returns usage statistics for a specific instance
func GetInstanceUsage() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check authentication but we don't need to use userID in this example
		_, err := middleware.GetUserIDFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		instanceID := c.Param("instanceId")
		if instanceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Instance ID is required"})
			return
		}

		// TODO: Implement actual usage statistics retrieval
		// This is a placeholder
		usageStats := map[string]interface{}{
			"cpu": map[string]interface{}{
				"current": 0.2,
				"limit":   1.0,
				"unit":    "cores",
			},
			"memory": map[string]interface{}{
				"current": 128,
				"limit":   512,
				"unit":    "MB",
			},
			"storage": map[string]interface{}{
				"current": 0.5,
				"limit":   1.0,
				"unit":    "GB",
			},
			"network": map[string]interface{}{
				"in":  10.5,
				"out": 5.2,
				"unit": "MB",
			},
		}

		c.JSON(http.StatusOK, usageStats)
	}
} 