package routes

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/launchstack/backend/config"
	"github.com/sirupsen/logrus"
	svix "github.com/svix/svix-webhooks/go"
)

// WebhookHandler handles incoming Clerk webhook events
func WebhookHandler(cfg *config.Config, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Infof("Received webhook request to path: %s", c.Request.URL.Path)
		
		// Log headers for debugging
		logger.Info("Request headers:")
		for key, values := range c.Request.Header {
			logger.Infof("  %s: %s", key, strings.Join(values, ", "))
		}
		
		// Read and store the request body so we can verify signature and then process it
		var buf bytes.Buffer
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Errorf("Error reading webhook body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		
		// Log the request body for debugging
		logger.Infof("Request body: %s", string(body))
		
		// Restore the request body for further processing
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		buf.Write(body)
		
		// Skip signature verification if webhook secret is not configured
		// or if we're in development mode and request is from localhost
		skipVerification := cfg.Clerk.WebhookSecret == ""
		
		// Check if request is from local testing
		clientIP := c.ClientIP()
		isLocalRequest := clientIP == "127.0.0.1" || clientIP == "::1" || strings.HasPrefix(clientIP, "192.168.") || strings.HasPrefix(clientIP, "10.")
		
		// Get the forwarded IP if available
		forwardedFor := c.GetHeader("X-Forwarded-For")
		if forwardedFor != "" {
			logger.Infof("Request from forwarded IP: %s", forwardedFor)
		}
		
		// Check webhook signature using svix library
		verified := false
		if !skipVerification && cfg.Clerk.WebhookSecret != "whsec_replace_with_your_clerk_webhook_secret" {
			wh, err := svix.NewWebhook(cfg.Clerk.WebhookSecret)
			if err != nil {
				logger.Errorf("Error creating svix webhook verifier: %v", err)
			} else {
				err = wh.Verify(body, c.Request.Header)
				if err != nil {
					logger.Errorf("Webhook verification failed: %v", err)
				} else {
					logger.Info("Webhook signature verified successfully using svix library")
					verified = true
				}
			}
		}
		
		// Development mode bypass for testing
		if !verified && cfg.Server.Environment == "development" && (isLocalRequest || skipVerification) {
			logger.Warn("Development mode: Bypassing signature verification")
			skipVerification = true
		}
		
		// In development mode, continue despite verification failure
		if !verified && !skipVerification {
			if cfg.Server.Environment == "development" {
				logger.Warn("Development mode: Processing webhook despite signature verification failure")
				logger.Warn("Update CLERK_WEBHOOK_SECRET in your .env file with the actual webhook secret from Clerk dashboard")
			} else {
				logger.Error("Invalid webhook signature - rejecting request")
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
				return
			}
		}
		
		// Process the webhook event
		if err := ProcessWebhookEvent(body, logger); err != nil {
			logger.Errorf("Error processing webhook event: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process webhook"})
			return
		}
		
		// Return a success response
		c.JSON(http.StatusOK, gin.H{"message": "Webhook processed successfully"})
	}
}

// RegisterClerkWebhookRoutes registers all webhook routes
func RegisterClerkWebhookRoutes(router *gin.Engine, cfg *config.Config, logger *logrus.Logger) {
	logger.Info("Registering Clerk webhook routes")
	
	// Register the webhook handler under the /api/v1/webhooks/clerk path
	webhookGroup := router.Group("/api/v1/webhooks")
	{
		webhookGroup.POST("/clerk", WebhookHandler(cfg, logger))
	}
	
	logger.Info("Clerk webhook routes registered successfully")
} 