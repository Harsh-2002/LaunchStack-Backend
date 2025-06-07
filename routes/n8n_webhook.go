package routes

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// N8nWebhookRequest represents a webhook request from n8n
type N8nWebhookRequest struct {
	Event       string                 `json:"event"`
	InstanceID  string                 `json:"instanceId"`
	Payload     map[string]interface{} `json:"payload"`
	WorkflowID  string                 `json:"workflowId,omitempty"`
	ExecutionID string                 `json:"executionId,omitempty"`
}

// N8nWebhook handles webhook events from n8n instances
func N8nWebhook(webhookSecret string, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.WithError(err).Error("Failed to read n8n webhook body")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		// Restore the request body for binding
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Verify webhook signature if secret is provided
		if webhookSecret != "" {
			signature := c.GetHeader("X-N8N-Signature")
			if signature == "" {
				logger.Error("Missing n8n webhook signature")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing signature"})
				return
			}

			// Calculate expected signature
			h := hmac.New(sha256.New, []byte(webhookSecret))
			h.Write(body)
			expectedSignature := hex.EncodeToString(h.Sum(nil))

			// Compare signatures
			if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
				logger.Error("Invalid n8n webhook signature")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
				return
			}
		}

		// Parse webhook request
		var webhook N8nWebhookRequest
		if err := c.ShouldBindJSON(&webhook); err != nil {
			logger.WithError(err).Error("Failed to parse n8n webhook")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload"})
			return
		}

		// Handle different event types
		switch webhook.Event {
		case "workflow.started":
			handleWorkflowStarted(c, webhook, logger)
		case "workflow.completed":
			handleWorkflowCompleted(c, webhook, logger)
		case "workflow.failed":
			handleWorkflowFailed(c, webhook, logger)
		case "instance.status":
			handleInstanceStatus(c, webhook, logger)
		default:
			// Acknowledge receipt of the webhook but take no action
			logger.WithField("event", webhook.Event).Info("Received unhandled n8n webhook event")
			c.JSON(http.StatusOK, gin.H{"status": "acknowledged"})
		}
	}
}

// handleWorkflowStarted handles workflow.started events
func handleWorkflowStarted(c *gin.Context, webhook N8nWebhookRequest, logger *logrus.Logger) {
	// In a real implementation, you would track workflow executions in your database
	logger.WithFields(logrus.Fields{
		"instance_id":  webhook.InstanceID,
		"workflow_id":  webhook.WorkflowID,
		"execution_id": webhook.ExecutionID,
	}).Info("Workflow started")

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// handleWorkflowCompleted handles workflow.completed events
func handleWorkflowCompleted(c *gin.Context, webhook N8nWebhookRequest, logger *logrus.Logger) {
	// In a real implementation, you would update workflow execution status in your database
	logger.WithFields(logrus.Fields{
		"instance_id":  webhook.InstanceID,
		"workflow_id":  webhook.WorkflowID,
		"execution_id": webhook.ExecutionID,
	}).Info("Workflow completed")

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// handleWorkflowFailed handles workflow.failed events
func handleWorkflowFailed(c *gin.Context, webhook N8nWebhookRequest, logger *logrus.Logger) {
	// In a real implementation, you would update workflow execution status in your database
	// and potentially trigger notifications
	logger.WithFields(logrus.Fields{
		"instance_id":  webhook.InstanceID,
		"workflow_id":  webhook.WorkflowID,
		"execution_id": webhook.ExecutionID,
		"error":        webhook.Payload["error"],
	}).Warn("Workflow failed")

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// handleInstanceStatus handles instance.status events
func handleInstanceStatus(c *gin.Context, webhook N8nWebhookRequest, logger *logrus.Logger) {
	// In a real implementation, you would update instance status in your database
	status, ok := webhook.Payload["status"].(string)
	if !ok {
		logger.Error("Missing status in instance.status event")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing status"})
		return
	}

	logger.WithFields(logrus.Fields{
		"instance_id": webhook.InstanceID,
		"status":      status,
	}).Info("Instance status update")

	// Update instance status in database if applicable
	if webhook.InstanceID != "" {
		// Placeholder for database update
		// In a real implementation, you would look up the instance and update its status
		
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing instance ID"})
	}
} 