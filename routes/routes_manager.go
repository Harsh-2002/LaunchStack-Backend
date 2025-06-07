package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/launchstack/backend/config"
	"github.com/launchstack/backend/container"
	"github.com/launchstack/backend/db"
	"github.com/launchstack/backend/middleware"
	"github.com/sirupsen/logrus"
)

// RegisterAllRoutes sets up all the API routes
func RegisterAllRoutes(
	router *gin.Engine,
	cfg *config.Config,
	containerManager *container.Manager,
	logger *logrus.Logger,
) {
	// Check if database connection is initialized
	if db.DB == nil {
		logger.Error("Database connection not initialized! Routes requiring database will not work properly.")
	} else {
		logger.Info("Database connection initialized and ready")
	}

	// API group with versioning
	api := router.Group("/api/v1")

	// Public routes
	public := api.Group("/")
	{
		public.GET("/health", HealthCheck(cfg))
	}

	// Protected routes requiring authentication
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg.Clerk.SecretKey, logger, cfg))
	{
		// Instance management
		instanceRoutes := protected.Group("/instances")
		{
			instanceRoutes.GET("/", GetInstances(containerManager))
			instanceRoutes.POST("/", CreateInstance(containerManager))
			instanceRoutes.GET("/:id", GetInstance(containerManager))
			instanceRoutes.PUT("/:id", UpdateInstance(containerManager))
			instanceRoutes.DELETE("/:id", DeleteInstance(containerManager))
			instanceRoutes.POST("/:id/start", StartInstance(containerManager))
			instanceRoutes.POST("/:id/stop", StopInstance(containerManager))
			instanceRoutes.POST("/:id/restart", RestartInstance(containerManager))
		}

		// User management
		userRoutes := protected.Group("/users")
		{
			userRoutes.GET("/me", GetCurrentUser())
			userRoutes.PUT("/me", UpdateCurrentUser())
		}

		// Only register payment routes if payments are not disabled
		if !cfg.PayPal.DisablePayments {
			// Payment management
			paymentRoutes := protected.Group("/payments")
			{
				paymentRoutes.GET("/", GetPayments)
				paymentRoutes.POST("/checkout", CreateCheckoutSession)
				paymentRoutes.GET("/subscriptions", GetSubscriptions)
				paymentRoutes.POST("/subscriptions/:id/cancel", CancelSubscription)
			}
		}

		// Usage statistics
		usageRoutes := protected.Group("/usage")
		{
			usageRoutes.GET("/", GetUsageStats())
			usageRoutes.GET("/:instanceId", GetInstanceUsage())
		}
	}

	// Webhook routes (not requiring the same authentication)
	webhooks := api.Group("/webhooks")
	{
		// Only register PayPal webhooks if payments are not disabled
		if !cfg.PayPal.DisablePayments {
			webhooks.POST("/paypal", PayPalWebhook)
		}
		
		// Register Clerk webhooks
		logger.Info("Registering Clerk webhook routes...")
		if cfg.Clerk.WebhookSecret == "" {
			logger.Warn("Clerk webhook secret is not set - signature verification will be disabled")
		}
		
		// Register webhook routes through the dedicated function
		// This function handles all clerk webhook endpoints
		RegisterClerkWebhookRoutes(router, cfg, logger)
		
		// n8n webhook for workflow and instance events
		// Only register if the secret is set
		n8nSecret := cfg.N8N.WebhookSecret
		if n8nSecret != "" {
			webhooks.POST("/n8n", N8nWebhook(n8nSecret, logger))
		}
	}
} 