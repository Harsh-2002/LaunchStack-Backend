package main

import (
	"fmt"
	
	"github.com/joho/godotenv"
	"github.com/launchstack/backend/config"
	"github.com/launchstack/backend/db"
	"github.com/launchstack/backend/models"
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
	
	// Get all users
	var users []models.User
	result := db.DB.Find(&users)
	if result.Error != nil {
		logger.Fatalf("Failed to query users: %v", result.Error)
	}
	
	// Print user details
	fmt.Printf("Found %d users:\n\n", len(users))
	for i, user := range users {
		fmt.Printf("User #%d:\n", i+1)
		fmt.Printf("  ID: %s\n", user.ID)
		fmt.Printf("  Clerk User ID: %s\n", user.ClerkUserID)
		fmt.Printf("  Email: %s\n", user.Email)
		fmt.Printf("  Name: %s %s\n", user.FirstName, user.LastName)
		fmt.Printf("  Plan: %s\n", user.Plan)
		fmt.Printf("  Subscription Status: %s\n", user.SubscriptionStatus)
		fmt.Printf("  Created At: %s\n", user.CreatedAt)
		fmt.Printf("  Updated At: %s\n", user.UpdatedAt)
		fmt.Printf("  Deleted At: %v\n", user.DeletedAt)
		fmt.Println()
	}
} 