package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/launchstack/backend/config"
	"github.com/launchstack/backend/container"
	"github.com/launchstack/backend/db"
	"github.com/launchstack/backend/middleware"
	"github.com/launchstack/backend/routes"
	"github.com/sirupsen/logrus"
)

// getCORSOrigins gets the CORS origins directly from the environment variable
func getCORSOrigins(logger *logrus.Logger) []string {
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		// Default to localhost if not set
		logger.Warn("CORS_ORIGINS environment variable not set, defaulting to localhost:3000")
		return []string{"http://localhost:3000"}
	}
	
	// Split by comma and trim spaces
	origins := strings.Split(corsOrigins, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	
	logger.Infof("CORS origins loaded from environment: %v", origins)
	return origins
}

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		logger.Warning("Error loading .env file, using environment variables")
	}
	
	// Initialize configuration
	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Set log level based on configuration
	logLevel, err := logrus.ParseLevel(cfg.Monitoring.LogLevel)
	if err != nil {
		logger.Warnf("Invalid log level: %s, defaulting to info", cfg.Monitoring.LogLevel)
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)
	
	// Initialize database connection
	logger.Info("Connecting to database...")
	err = db.Initialize(cfg.Database.URL)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	logger.Info("Successfully connected to database")
	
	// Run database migrations
	logger.Info("Running database migrations...")
	err = db.RunMigrations()
	if err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}
	logger.Info("Database migrations completed successfully")
	
	// Create mock container manager
	logger.Info("Creating mock container manager...")
	containerManager := container.NewManager(logger)
	logger.Info("Mock container manager created successfully")
	
	// Get CORS origins directly from environment
	corsOrigins := getCORSOrigins(logger)
	
	// Initialize router
	router := gin.Default()
	
	// Add middleware
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.CORSMiddleware(corsOrigins))
	
	// Health check endpoint
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"version": "0.1.0",
			"environment": cfg.Server.Environment,
			"payment_gateway": "PayPal",
			"cors_origins": corsOrigins,
		})
	})
	
	// Register all routes
	logger.Info("Registering routes...")
	routes.RegisterRoutes(router, cfg, containerManager, logger)
	
	// Register Clerk webhook routes
	logger.Info("Registering Clerk webhook routes...")
	routes.RegisterClerkWebhookRoutes(router, cfg, logger)
	
	// Log all registered routes
	for _, routeInfo := range router.Routes() {
		logger.Infof("Registered route: %s %s", routeInfo.Method, routeInfo.Path)
	}
	
	// Start server
	port := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Infof("Starting server on port %s...", port)
	if err := router.Run("0.0.0.0" + port); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
} 