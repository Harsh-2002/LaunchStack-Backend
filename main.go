package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/launchstack/backend/config"
	"github.com/launchstack/backend/container"
	"github.com/launchstack/backend/db"
	"github.com/launchstack/backend/middleware"
	"github.com/launchstack/backend/models"
	"github.com/launchstack/backend/routes"
	"github.com/sirupsen/logrus"
)

// Docker client initialization is now handled by container.NewDockerClient

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

// initializeDatabase initializes the database connection and runs migrations
func initializeDatabase(cfg *config.Config, logger *logrus.Logger) error {
	// Initialize database connection with retries
	maxRetries := 5
	retryDelay := time.Second * 3
	
	var err error
	for i := 0; i < maxRetries; i++ {
		logger.Infof("Connecting to database (attempt %d/%d)...", i+1, maxRetries)
		err = db.Initialize(cfg.Database.URL)
		if err == nil {
			logger.Info("Successfully connected to database")
			break
		}
		
		logger.Warnf("Failed to connect to database: %v", err)
		if i < maxRetries-1 {
			logger.Infof("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}
	
	if err != nil {
		return fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}
	
	// Run database migrations using the smart migration system
	logger.Info("Checking if migrations need to be run...")
	if err = db.RunMigrationsWithLogger(logger); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	
	// Verify migrations by checking if tables exist
	logger.Info("Verifying database schema...")
	if err = verifyDatabaseSchema(logger); err != nil {
		return fmt.Errorf("database schema verification failed: %w", err)
	}
	
	// Debug log the configuration
	logger.Infof("DisablePayments: %v, Environment: %s", cfg.PayPal.DisablePayments, cfg.Server.Environment)
	
	// Note: We're using real JWT authentication but still keeping payments disabled
	logger.Info("JWT authentication enabled - development user bypass disabled")
	
	logger.Info("Database initialized successfully")
	return nil
}

// verifyDatabaseSchema verifies that the database schema was properly created
func verifyDatabaseSchema(logger *logrus.Logger) error {
	// Check if we can access the users table
	var user models.User
	result := db.DB.First(&user)
	if result.Error != nil && !strings.Contains(result.Error.Error(), "record not found") {
		// If the error is not "record not found", there might be a schema issue
		return fmt.Errorf("failed to query users table: %v", result.Error)
	}
	
	// We either found a user or got "record not found" which means the table exists
	logger.Info("Users table verified")
	return nil
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
	// Temporarily set to Debug level to see more detailed logs
	logLevel = logrus.DebugLevel
	logger.Infof("Setting log level to DEBUG for detailed request logging")
	logger.SetLevel(logLevel)
	
	// Initialize database and run migrations
	if err := initializeDatabase(cfg, logger); err != nil {
		logger.Fatalf("Database initialization failed: %v", err)
	}
	
	// Initialize Docker client
	var dockerClient container.DockerClient
	logger.Info("Initializing Docker client...")
	dockerClient, err = container.NewDockerClient(cfg.Docker.Host)
	if err != nil {
		logger.Warnf("Failed to initialize Docker client: %v", err)
		logger.Info("Continuing without Docker support...")
	} else {
		logger.Info("Docker client initialized successfully")
		
		// List running containers
		containers, err := dockerClient.ContainerList(context.Background(), types.ContainerListOptions{})
		if err != nil {
			logger.Warnf("Failed to list containers: %v", err)
		} else {
			logger.Infof("Found %d running containers", len(containers))
		}
	}
	
	// Get CORS origins directly from environment
	corsOrigins := getCORSOrigins(logger)
	
	// Create container manager based on the configuration
	var containerManager container.Manager
	if cfg.Docker.Host != "" {
		// Create Docker client
		dockerClient, err := container.NewDockerClient(cfg.Docker.Host)
		if err != nil {
			logger.WithError(err).Fatal("Failed to create Docker client")
		}
		
		// Create Docker container manager
		containerManager = container.NewManager(dockerClient, cfg, logger)
	} else {
		// Fall back to mock container manager
		containerManager = container.NewMockManager(logger, cfg)
	}
	
	// Start resource monitoring in a background goroutine
	go func() {
		logger.Infof("Starting resource usage monitoring every %v", cfg.Monitoring.Interval)
		ticker := time.NewTicker(cfg.Monitoring.Interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				// Get all active instances
				var instances []models.Instance
				if result := db.DB.Where("status != ?", models.StatusDeleted).Find(&instances); result.Error != nil {
					logger.WithError(result.Error).Error("Failed to fetch instances for resource monitoring")
					continue
				}
				
				// Collect stats for each instance
				for _, instance := range instances {
					go func(inst models.Instance) {
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()
						
						_, err := containerManager.GetInstanceStats(ctx, inst.ID)
						if err != nil {
							logger.WithFields(logrus.Fields{
								"instance_id": inst.ID,
								"error":      err.Error(),
							}).Warn("Failed to collect stats for instance")
						}
					}(instance)
				}
			}
		}
	}()
	
	// Initialize router
	router := gin.Default()
	
	// Add middleware
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.CORSMiddleware(corsOrigins))
	router.Use(middleware.AuthMiddleware(cfg.Clerk.SecretKey, logger, cfg))
	
	// Log configuration for debugging
	logger.WithFields(logrus.Fields{
		"environment":      cfg.Server.Environment,
		"disable_payments": cfg.PayPal.DisablePayments,
		"auth_enabled":     true,
		"dev_user_bypass":  false,
	}).Info("Server configuration - using real JWT authentication")
	
	// Debug middleware configuration
	logger.WithFields(logrus.Fields{
		"ContextMiddleware": true,
		"CORSMiddleware":    true,
		"AuthMiddleware":    true,
		"ContainerManager":  containerManager != nil,
	}).Info("Debug middleware configuration before registering routes")
	
	routes.RegisterAllRoutes(router, cfg, containerManager, logger)
	
	// Register Clerk webhook routes
	routes.RegisterClerkWebhookRoutes(router, cfg, logger)
	
	// Register mock payment routes if in development mode with payments disabled
	if cfg.PayPal.DisablePayments && cfg.Server.Environment == "development" {
		logger.Info("Registering mock payment routes for development mode")
		routes.RegisterMockPaymentRoutes(router, logger)
	}
	
	// Log all registered routes
	for _, routeInfo := range router.Routes() {
		logger.Infof("Registered route: %s %s", routeInfo.Method, routeInfo.Path)
	}
	
	// Start server
	port := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Infof("Starting server on port %s...", port)
	if err := router.Run(port); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
} 