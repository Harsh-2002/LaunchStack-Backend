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

// ClerkWebhookHandler handles Clerk webhook events
func ClerkWebhookHandler(cfg *config.Config, logger *logrus.Logger) gin.HandlerFunc {
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
		isLocalRequest := strings.Contains(c.ClientIP(), "127.0.0.1") || strings.Contains(c.ClientIP(), "::1")
		skipVerification := cfg.Clerk.WebhookSecret == ""
		
		// Get the forwarded IP if available
		forwardedFor := c.GetHeader("X-Forwarded-For")
		if forwardedFor != "" {
			logger.Infof("Request from forwarded IP: %s", forwardedFor)
		}
		
		// Check webhook signature using svix library
		verified := false
		if !skipVerification && cfg.Clerk.WebhookSecret != "" {
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
		
		// For local development environments, allow bypassing verification
		if !verified && cfg.Server.Environment == "development" && (isLocalRequest || skipVerification) {
			logger.Warn("Bypassing webhook signature verification for local development")
			verified = true
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