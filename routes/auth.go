package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/launchstack/backend/config"
	"github.com/sirupsen/logrus"
)

// ClerkWebhookHandler handles Clerk webhook events
func ClerkWebhookHandler(cfg *config.Config, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Basic implementation - will be expanded later
		logger.Debug("Received Clerk webhook event")
		c.JSON(http.StatusOK, gin.H{"status": "received"})
	}
} 