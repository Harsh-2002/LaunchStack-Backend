package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/launchstack/backend/config"
	"github.com/launchstack/backend/container"
	"github.com/sirupsen/logrus"
)

// RegisterAllRoutes registers all routes
func RegisterAllRoutes(router *gin.Engine, cfg *config.Config, containerManager container.Manager, logger *logrus.Logger) {
	// Register auth routes
	RegisterAuthRoutes(router, cfg, logger)
	
	// Register instance routes
	RegisterInstanceRoutes(router, cfg, containerManager, logger)
	
	// Register user routes
	RegisterUserRoutes(router, cfg, logger)
	
	// Register health check routes - redirect old paths to new /api/v1/ path
	router.GET("/health", func(c *gin.Context) {
		c.Redirect(301, "/api/v1/health")
	})
	
	// Standard v1 health check endpoint
	router.GET("/api/v1/health", HealthCheckHandler(cfg, logger))
	router.GET("/api/v1/health/", HealthCheckHandler(cfg, logger))
}

// RegisterAuthRoutes registers authentication-related routes
func RegisterAuthRoutes(router *gin.Engine, cfg *config.Config, logger *logrus.Logger) {
	// Register redirects for old routes
	oldAuthRoutes := router.Group("/api/auth")
	oldAuthRoutes.POST("/webhook", func(c *gin.Context) {
		c.Redirect(301, "/api/v1/auth/webhook")
	})
	
	// Register v1 auth routes
	v1AuthRoutes := router.Group("/api/v1/auth")
	v1AuthRoutes.POST("/webhook", ClerkWebhookHandler(cfg, logger))
	v1AuthRoutes.POST("/webhook/", ClerkWebhookHandler(cfg, logger))
}

// RegisterUserRoutes registers user-related routes
func RegisterUserRoutes(router *gin.Engine, cfg *config.Config, logger *logrus.Logger) {
	// Register redirects for old routes
	oldUserRoutes := router.Group("/api/users")
	oldUserRoutes.GET("/me", func(c *gin.Context) {
		c.Redirect(301, "/api/v1/users/me")
	})
	oldUserRoutes.PUT("/me", func(c *gin.Context) {
		c.Redirect(301, "/api/v1/users/me")
	})
	
	// Register v1 user routes
	v1UserRoutes := router.Group("/api/v1/users")
	v1UserRoutes.GET("/me", GetCurrentUserHandler)
	v1UserRoutes.GET("/me/", GetCurrentUserHandler)
	v1UserRoutes.PUT("/me", UpdateCurrentUserHandler)
	v1UserRoutes.PUT("/me/", UpdateCurrentUserHandler)
}

// RegisterInstanceRoutes registers instance-related routes
func RegisterInstanceRoutes(router *gin.Engine, cfg *config.Config, containerManager container.Manager, logger *logrus.Logger) {
	// Register redirects for old routes
	oldInstanceRoutes := router.Group("/api/instances")
	oldInstanceRoutes.GET("", func(c *gin.Context) {
		c.Redirect(301, "/api/v1/instances")
	})
	oldInstanceRoutes.POST("", func(c *gin.Context) {
		c.Redirect(301, "/api/v1/instances")
	})
	oldInstanceRoutes.GET("/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.Redirect(301, "/api/v1/instances/"+id)
	})
	oldInstanceRoutes.DELETE("/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.Redirect(301, "/api/v1/instances/"+id)
	})
	oldInstanceRoutes.POST("/:id/start", func(c *gin.Context) {
		id := c.Param("id")
		c.Redirect(301, "/api/v1/instances/"+id+"/start")
	})
	oldInstanceRoutes.POST("/:id/stop", func(c *gin.Context) {
		id := c.Param("id")
		c.Redirect(301, "/api/v1/instances/"+id+"/stop")
	})
	oldInstanceRoutes.POST("/:id/restart", func(c *gin.Context) {
		id := c.Param("id")
		c.Redirect(301, "/api/v1/instances/"+id+"/restart")
	})
	oldInstanceRoutes.GET("/:id/stats", func(c *gin.Context) {
		id := c.Param("id")
		c.Redirect(301, "/api/v1/instances/"+id+"/stats")
	})
	oldInstanceRoutes.GET("/:id/stats/history", func(c *gin.Context) {
		id := c.Param("id")
		c.Redirect(301, "/api/v1/instances/"+id+"/stats/history")
	})
	
	// Register v1 instance routes
	v1InstanceRoutes := router.Group("/api/v1/instances")
	v1InstanceRoutes.Use(ContainerManagerMiddleware(containerManager))
	
	// Register all v1 instance routes with proper handler functions
	// Make sure to handle both with and without trailing slashes
	v1InstanceRoutes.GET("", GetInstances(containerManager))
	v1InstanceRoutes.GET("/", GetInstances(containerManager))
	v1InstanceRoutes.POST("", CreateInstance(containerManager))
	v1InstanceRoutes.POST("/", CreateInstance(containerManager))
	v1InstanceRoutes.GET("/:id", GetInstance(containerManager))
	v1InstanceRoutes.GET("/:id/", GetInstance(containerManager))
	v1InstanceRoutes.DELETE("/:id", DeleteInstance(containerManager))
	v1InstanceRoutes.DELETE("/:id/", DeleteInstance(containerManager))
	v1InstanceRoutes.POST("/:id/start", StartInstance(containerManager))
	v1InstanceRoutes.POST("/:id/start/", StartInstance(containerManager))
	v1InstanceRoutes.POST("/:id/stop", StopInstance(containerManager))
	v1InstanceRoutes.POST("/:id/stop/", StopInstance(containerManager))
	v1InstanceRoutes.POST("/:id/restart", RestartInstance(containerManager))
	v1InstanceRoutes.POST("/:id/restart/", RestartInstance(containerManager))
	v1InstanceRoutes.GET("/:id/stats", GetInstanceStats(containerManager))
	v1InstanceRoutes.GET("/:id/stats/", GetInstanceStats(containerManager))
	
	// Add the historical stats endpoint with the path expected by frontend
	v1InstanceRoutes.GET("/:id/stats/history", GetInstanceHistoricalStats())
	v1InstanceRoutes.GET("/:id/stats/history/", GetInstanceHistoricalStats())
} 