package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// DEV_USER_ID is the UUID used for the development user
const DEV_USER_ID = "f2814e7b-75a0-44d4-b345-e5ef5a84aab3"

// Payment status constants
const (
	PaymentStatusPending   = "pending"
	PaymentStatusCompleted = "completed"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"
)

// RegisterMockPaymentRoutes registers mock payment routes for development mode
func RegisterMockPaymentRoutes(router *gin.Engine, logger *logrus.Logger) {
	logger.Info("Registering mock payment routes for development mode")

	// API group with versioning
	api := router.Group("/api/v1")

	// Payment routes with a simple dev auth middleware
	paymentRoutes := api.Group("/payments")
	paymentRoutes.Use(func(c *gin.Context) {
		// Add the development user ID to the context
		userID, _ := uuid.Parse(DEV_USER_ID)
		c.Set("userID", userID)
		c.Set("userEmail", "dev@launchstack.io")
		c.Next()
	})
	
	{
		paymentRoutes.GET("", MockGetPayments)
		paymentRoutes.GET("/", MockGetPayments)
		
		paymentRoutes.POST("/checkout", MockCreateCheckoutSession)
		paymentRoutes.POST("/checkout/", MockCreateCheckoutSession)
		
		paymentRoutes.GET("/subscriptions", MockGetSubscriptions)
		paymentRoutes.GET("/subscriptions/", MockGetSubscriptions)
		
		paymentRoutes.POST("/subscriptions/:id/cancel", MockCancelSubscription)
		paymentRoutes.POST("/subscriptions/:id/cancel/", MockCancelSubscription)
	}

	// Mock webhook route
	webhooks := api.Group("/webhooks")
	{
		webhooks.POST("/paypal", MockPayPalWebhook)
		webhooks.POST("/paypal/", MockPayPalWebhook)
	}

	logger.Info("Mock payment routes registered successfully")
}

// MockGetPayments returns mock payment history
func MockGetPayments(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Create mock payment data
	payments := []gin.H{
		{
			"id":          uuid.New().String(),
			"user_id":     userID.(uuid.UUID).String(),
			"amount":      500, // $5.00 for Pro plan
			"currency":    "usd",
			"status":      PaymentStatusCompleted,
			"description": "Subscription to pro plan",
			"created_at":  time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"payments": payments,
	})
}

// MockCreateCheckoutSession creates a mock checkout session
func MockCreateCheckoutSession(c *gin.Context) {
	// Parse request body
	var req struct {
		Plan       string `json:"plan"`
		SuccessURL string `json:"success_url"`
		CancelURL  string `json:"cancel_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Return mock checkout URL
	c.JSON(http.StatusOK, gin.H{
		"checkout_url": req.SuccessURL + "?success=true",
		"order_id":     "MOCK-ORDER-" + uuid.New().String(),
	})
}

// MockGetSubscriptions returns mock subscription data
func MockGetSubscriptions(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Mock active subscription
	subscription := gin.H{
		"id":          "MOCK-SUB-" + uuid.New().String(),
		"user_id":     userID.(uuid.UUID).String(),
		"status":      "ACTIVE",
		"plan":        "pro",
		"start_date":  time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339),
		"end_date":    time.Now().Add(335 * 24 * time.Hour).Format(time.RFC3339),
		"auto_renew":  true,
		"amount":      500, // $5.00 per month
		"currency":    "usd",
		"description": "Pro Plan Subscription",
	}

	c.JSON(http.StatusOK, gin.H{
		"subscription": subscription,
		"has_active":   true,
	})
}

// MockCancelSubscription mocks canceling a subscription
func MockCancelSubscription(c *gin.Context) {
	// Get subscription ID from path
	subscriptionID := c.Param("id")
	if subscriptionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscription ID is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Subscription canceled successfully",
		"end_date":  time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
	})
}

// MockPayPalWebhook handles mock PayPal webhooks
func MockPayPalWebhook(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
} 