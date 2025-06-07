package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/launchstack/backend/db"
	"github.com/launchstack/backend/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// WebhookEvent represents the common structure of Clerk webhook events
type WebhookEvent struct {
	Data        json.RawMessage `json:"data"`
	Object      string          `json:"object"`
	Type        string          `json:"type"`
}

// UserData represents user data in a Clerk webhook
type UserData struct {
	ID              string                `json:"id"`
	FirstName       string                `json:"first_name"`
	LastName        string                `json:"last_name"`
	EmailAddresses  []ClerkEmailAddress   `json:"email_addresses"`
	CreatedAt       int64                 `json:"created_at"`
	UpdatedAt       int64                 `json:"updated_at"`
	LastSignInAt    int64                 `json:"last_sign_in_at"`
	PrimaryEmailID  string                `json:"primary_email_address_id"`
	ProfileImageURL string                `json:"profile_image_url"`
	// Add other fields as needed
}

// ClerkEmailAddress represents an email address in Clerk user data
type ClerkEmailAddress struct {
	ID           string `json:"id"`
	EmailAddress string `json:"email_address"`
	Verification struct {
		Status   string `json:"status"`
		Strategy string `json:"strategy"`
	} `json:"verification"`
}

// ProcessWebhookEvent processes different Clerk webhook events
func ProcessWebhookEvent(eventBody []byte, logger *logrus.Logger) error {
	var event WebhookEvent
	if err := json.Unmarshal(eventBody, &event); err != nil {
		logger.Errorf("Failed to parse webhook event: %v", err)
		return err
	}

	logger.Infof("Processing webhook event of type: %s", event.Type)
	
	// Verify database connection is initialized
	if db.DB == nil {
		err := fmt.Errorf("database connection is not initialized")
		logger.Errorf("Database error: %v", err)
		return err
	}
	
	// Check database connection by pinging
	sqlDB, err := db.DB.DB()
	if err != nil {
		logger.Errorf("Failed to get SQL DB: %v", err)
		return err
	}
	
	if err := sqlDB.Ping(); err != nil {
		logger.Errorf("Database ping failed: %v", err)
		return err
	}
	
	logger.Info("Database connection verified")

	// Log the entire event for debugging
	prettyJSON, _ := json.MarshalIndent(event, "", "  ")
	logger.Infof("Full event: %s", string(prettyJSON))
	
	// Handle different event types
	switch event.Type {
	case "user.created":
		// Real Clerk webhooks have the user data inside the "data" field
		return handleUserCreated(event.Data, logger)
	case "user.updated":
		return handleUserUpdated(event.Data, logger)
	case "user.deleted":
		return handleUserDeleted(event.Data, logger)
	default:
		logger.Infof("Unhandled event type: %s", event.Type)
		return nil
	}
}

// handleUserCreated processes user.created events
func handleUserCreated(data json.RawMessage, logger *logrus.Logger) error {
	var userData UserData
	if err := json.Unmarshal(data, &userData); err != nil {
		logger.Errorf("Failed to parse user data: %v", err)
		
		// Try to extract the user ID at least for logging
		var rawData map[string]interface{}
		if err := json.Unmarshal(data, &rawData); err == nil {
			if id, ok := rawData["id"].(string); ok {
				logger.Infof("Extracted user ID from raw data: %s", id)
			}
		}
		
		// Log the raw data
		logger.Errorf("Raw data that failed to parse: %s", string(data))
		return err
	}

	// Log the raw user data for debugging
	rawData, _ := json.MarshalIndent(userData, "", "  ")
	logger.Infof("Raw user data: %s", string(rawData))

	// Find primary email address
	var primaryEmail string
	primaryEmailFound := false
	
	for _, email := range userData.EmailAddresses {
		logger.Infof("Checking email ID: %s vs primary ID: %s", email.ID, userData.PrimaryEmailID)
		if email.ID == userData.PrimaryEmailID {
			primaryEmail = email.EmailAddress
			primaryEmailFound = true
			break
		}
	}
	
	if !primaryEmailFound {
		// If we didn't find a matching ID, use the first email as fallback
		if len(userData.EmailAddresses) > 0 {
			primaryEmail = userData.EmailAddresses[0].EmailAddress
			logger.Warnf("Primary email ID not found, using first email: %s", primaryEmail)
		} else {
			logger.Errorf("No email addresses found for user %s", userData.ID)
			primaryEmail = fmt.Sprintf("unknown-%s@example.com", userData.ID)
		}
	}
	
	// Check if user already exists
	var existingUser models.User
	result := db.DB.Where("clerk_user_id = ?", userData.ID).First(&existingUser)
	if result.Error == nil {
		logger.Warnf("User with Clerk ID %s already exists, skipping creation", userData.ID)
		return nil
	}

	// Create a new user in our database
	user := &models.User{
		ID:            uuid.New(),
		ClerkUserID:   userData.ID,
		Email:         primaryEmail,
		FirstName:     userData.FirstName,
		LastName:      userData.LastName,
		Plan:          models.PlanFree, // Default to free plan
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Log the data we're about to save
	logger.Infof("Creating new user from Clerk: ID=%s, Email=%s, Name=%s %s", 
		user.ClerkUserID, user.Email, user.FirstName, user.LastName)

	// Save to database
	if err := db.DB.Create(user).Error; err != nil {
		logger.Errorf("Failed to create user in database: %v", err)
		return err
	}

	logger.Infof("Created new user in database: ID=%s, Clerk ID=%s", user.ID, user.ClerkUserID)
	return nil
}

// handleUserUpdated processes user.updated events
func handleUserUpdated(data json.RawMessage, logger *logrus.Logger) error {
	var userData UserData
	if err := json.Unmarshal(data, &userData); err != nil {
		logger.Errorf("Failed to parse user data: %v", err)
		return err
	}

	// Log the raw user data for debugging
	rawData, _ := json.MarshalIndent(userData, "", "  ")
	logger.Infof("Raw user data for update: %s", string(rawData))

	// Find primary email address
	var primaryEmail string
	for _, email := range userData.EmailAddresses {
		if email.ID == userData.PrimaryEmailID {
			primaryEmail = email.EmailAddress
			break
		}
	}
	
	if primaryEmail == "" && len(userData.EmailAddresses) > 0 {
		primaryEmail = userData.EmailAddresses[0].EmailAddress
		logger.Warnf("Primary email not found for update, using first email: %s", primaryEmail)
	}

	// Find the user in our database
	var user models.User
	result := db.DB.Where("clerk_user_id = ?", userData.ID).First(&user)
	if result.Error != nil {
		logger.Errorf("Failed to find user in database: %v", result.Error)
		
		// If user doesn't exist, create them (treating this as a user.created event)
		logger.Infof("User not found, creating instead: %s", userData.ID)
		return handleUserCreated(data, logger)
	}

	// Log the update
	logger.Infof("Updating user: ID=%s, New Email=%s, New Name=%s %s", 
		userData.ID, primaryEmail, userData.FirstName, userData.LastName)

	// Update user information
	user.Email = primaryEmail
	user.FirstName = userData.FirstName
	user.LastName = userData.LastName
	user.UpdatedAt = time.Now()

	// Save changes to database
	if err := db.DB.Save(&user).Error; err != nil {
		logger.Errorf("Failed to update user in database: %v", err)
		return err
	}

	logger.Infof("Updated user in database: ID=%s, Clerk ID=%s", user.ID, user.ClerkUserID)
	return nil
}

// handleUserDeleted processes user.deleted events
func handleUserDeleted(data json.RawMessage, logger *logrus.Logger) error {
	// For user.deleted events, the data structure is different
	var deletedUserData struct {
		ID      string `json:"id"`
		Deleted bool   `json:"deleted"`
		Object  string `json:"object"`
	}
	
	if err := json.Unmarshal(data, &deletedUserData); err != nil {
		logger.Errorf("Failed to parse deleted user data: %v", err)
		
		// Try to extract the user ID at least
		var rawData map[string]interface{}
		if err := json.Unmarshal(data, &rawData); err == nil {
			if id, ok := rawData["id"].(string); ok {
				logger.Infof("Extracted deleted user ID: %s", id)
				deletedUserData.ID = id
			}
		} else {
			return err
		}
	}
	
	if deletedUserData.ID == "" {
		logger.Error("No user ID found in deleted user data")
		return fmt.Errorf("missing user ID in deleted user data")
	}

	logger.Infof("Processing user deletion for Clerk user ID: %s", deletedUserData.ID)

	// Find the user in our database
	var user models.User
	result := db.DB.Where("clerk_user_id = ?", deletedUserData.ID).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// User doesn't exist in our database, which is fine - nothing to delete
			logger.Warnf("User with Clerk ID %s not found in database, nothing to delete", deletedUserData.ID)
			return nil
		}
		
		// For other database errors, return the error
		logger.Errorf("Failed to find user for deletion: %v", result.Error)
		return result.Error
	}

	// Soft delete the user
	if err := db.DB.Delete(&user).Error; err != nil {
		logger.Errorf("Failed to delete user from database: %v", err)
		return err
	}

	logger.Infof("Successfully deleted user: ID=%s, Clerk ID=%s", user.ID, user.ClerkUserID)
	return nil
} 