package db

import (
	"errors"

	"github.com/google/uuid"
	"github.com/launchstack/backend/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// FindUserByClerkID finds a user by their Clerk ID
func FindUserByClerkID(clerkID string) (models.User, error) {
	logger := getLogger()
	logger.WithField("clerk_user_id", clerkID).Info("Finding user by Clerk ID")
	
	var user models.User
	result := DB.Where("clerk_user_id = ?", clerkID).First(&user)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.WithField("clerk_user_id", clerkID).Warn("User not found with Clerk ID")
			return models.User{}, gorm.ErrRecordNotFound
		}
		logger.WithError(result.Error).Error("Database error when finding user by Clerk ID")
		return models.User{}, result.Error
	}
	
	logger.WithFields(logrus.Fields{
		"user_id":       user.ID,
		"clerk_user_id": user.ClerkUserID,
		"email":         user.Email,
	}).Info("Found user by Clerk ID")
	
	return user, nil
}

// GetUserByID gets a user by ID
func GetUserByID(id uuid.UUID) (models.User, error) {
	logger := getLogger()
	logger.WithField("user_id", id).Info("Getting user by ID")
	
	var user models.User
	result := DB.First(&user, "id = ?", id)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.WithField("user_id", id).Warn("User not found with ID")
			return models.User{}, gorm.ErrRecordNotFound
		}
		logger.WithError(result.Error).Error("Database error when finding user by ID")
		return models.User{}, result.Error
	}
	
	logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("Found user by ID")
	
	return user, nil
}

// CreateUser creates a new user
func CreateUser(user *models.User) error {
	logger := getLogger()
	logger.WithFields(logrus.Fields{
		"clerk_user_id": user.ClerkUserID,
		"email":         user.Email,
	}).Info("Creating new user")
	
	result := DB.Create(user)
	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to create user")
		return result.Error
	}
	
	logger.WithFields(logrus.Fields{
		"user_id":       user.ID,
		"clerk_user_id": user.ClerkUserID,
		"email":         user.Email,
	}).Info("Successfully created user")
	
	return nil
}

// UpdateUser updates an existing user
func UpdateUser(user *models.User) error {
	logger := getLogger()
	logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("Updating user")
	
	result := DB.Save(user)
	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to update user")
		return result.Error
	}
	
	logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("Successfully updated user")
	
	return nil
}

// DeleteUserByClerkID deletes a user by Clerk ID
func DeleteUserByClerkID(clerkID string) error {
	logger := getLogger()
	logger.WithField("clerk_user_id", clerkID).Info("Deleting user by Clerk ID")
	
	result := DB.Where("clerk_user_id = ?", clerkID).Delete(&models.User{})
	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to delete user")
		return result.Error
	}
	
	logger.WithField("clerk_user_id", clerkID).Info("Successfully deleted user")
	return nil
} 