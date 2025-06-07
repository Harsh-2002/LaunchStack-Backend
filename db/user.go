package db

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/launchstack/backend/models"
	"github.com/sirupsen/logrus"
)

// CreateUserIfNotExists creates a user if they don't already exist
func CreateUserIfNotExists(user models.User) error {
	logger := getLogger()
	
	// Check if user exists
	var existingUser models.User
	result := DB.Where("id = ?", user.ID).First(&existingUser)
	
	// If user exists, return
	if result.Error == nil {
		logger.WithField("user_id", user.ID).Info("User already exists")
		return nil
	}
	
	// Create user
	logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("Creating new user")
	
	if err := DB.Create(&user).Error; err != nil {
		logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"email":   user.Email,
			"error":   err.Error(),
		}).Error("Failed to create user")
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("Successfully created user")
	
	return nil
}

// CreateDevUser creates the development user if it doesn't exist
func CreateDevUser() error {
	logger := getLogger()
	logger.Info("Creating development user if needed")
	
	devUserID, _ := uuid.Parse("f2814e7b-75a0-44d4-b345-e5ef5a84aab3")
	devUser := models.User{
		ID:          devUserID,
		ClerkUserID: "dev-clerk-user",
		Email:       "dev@launchstack.io",
		FirstName:   "Development",
		LastName:    "User",
		Plan:        models.PlanPro,
	}
	
	return CreateUserIfNotExists(devUser)
} 