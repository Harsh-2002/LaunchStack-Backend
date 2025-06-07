package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/launchstack/backend/container"
)

// ContainerManagerMiddleware sets the container manager in the context
func ContainerManagerMiddleware(containerManager container.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if containerManager == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Container manager not available"})
			c.Abort()
			return
		}
		
		c.Set("container_manager", containerManager)
		c.Next()
	}
} 