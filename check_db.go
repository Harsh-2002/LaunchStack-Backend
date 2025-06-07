package main

import (
	"os"
	
	"github.com/joho/godotenv"
	"github.com/launchstack/backend/config"
	"github.com/launchstack/backend/db"
	"github.com/sirupsen/logrus"
)

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
	
	// Initialize database connection
	logger.Info("Connecting to database...")
	err = db.Initialize(cfg.Database.URL)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	logger.Info("Successfully connected to database")
	
	// Run migrations
	logger.Info("Running database migrations...")
	err = db.RunMigrations()
	if err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}
	logger.Info("Migrations completed successfully")
	
	// Run the SQL migration for PayPal integration
	logger.Info("Running PayPal integration SQL migration...")
	sqlFile, err := os.ReadFile("db/migrations/paypal_integration.sql")
	if err != nil {
		logger.Fatalf("Failed to read PayPal migration file: %v", err)
	}
	
	result := db.DB.Exec(string(sqlFile))
	if result.Error != nil {
		logger.Fatalf("Failed to execute PayPal migration: %v", result.Error)
	}
	
	logger.Info("PayPal integration migration completed successfully")
	
	// Test query to verify tables exist
	var count int64
	result = db.DB.Table("users").Count(&count)
	if result.Error != nil {
		logger.Fatalf("Failed to query users table: %v", result.Error)
	}
	logger.Infof("Users table exists and contains %d records", count)
	
	result = db.DB.Table("payments").Count(&count)
	if result.Error != nil {
		logger.Fatalf("Failed to query payments table: %v", result.Error)
	}
	logger.Infof("Payments table exists and contains %d records", count)
	
	logger.Info("Database check completed successfully")
} 