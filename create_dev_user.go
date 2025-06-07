package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/launchstack/backend/db"
	"github.com/launchstack/backend/models"
	"log"
	"os"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using environment variables")
	}
	
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}
	
	fmt.Println("Connecting to database...")
	err = db.Initialize(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	// Create fixed development user ID
	devUserID, _ := uuid.Parse("f2814e7b-75a0-44d4-b345-e5ef5a84aab3")
	
	// Create development user
	devUser := models.User{
		ID:          devUserID,
		ClerkUserID: "dev-clerk-user",
		Email:       "dev@launchstack.io",
		FirstName:   "Development",
		LastName:    "User",
		Plan:        models.PlanPro,
	}
	
	// Check if user exists
	var existingUser models.User
	result := db.DB.Where("id = ?", devUserID).First(&existingUser)
	
	if result.Error == nil {
		fmt.Println("Development user already exists!")
	} else {
		// Create the user
		err = db.DB.Create(&devUser).Error
		if err != nil {
			log.Fatalf("Failed to create development user: %v", err)
		}
		fmt.Println("Development user created successfully!")
	}
} 