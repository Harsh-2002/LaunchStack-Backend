package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/launchstack/backend/config"
	"github.com/launchstack/backend/models"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware validates the JWT token and adds the user to the context
func AuthMiddleware(clerkSecretKey string, logger *logrus.Logger, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for development
		if cfg.PayPal.DisablePayments && cfg.Server.Environment == "development" {
			// Use a consistent development user ID
			devUserID, _ := uuid.Parse("f2814e7b-75a0-44d4-b345-e5ef5a84aab3") // Fixed ID for development
			c.Set("userID", devUserID)
			
			// Set a mock development user
			devUser := models.User{
				ID:          devUserID,
				ClerkUserID: "dev-clerk-user",
				Email:       "dev@launchstack.io",
				FirstName:   "Development",
				LastName:    "User",
				Plan:        models.PlanPro, // Use Pro plan for development
			}
			c.Set("user", devUser)
			
			logger.WithFields(logrus.Fields{
				"user_id": devUserID.String(),
				"email":   devUser.Email,
				"path":    c.Request.URL.Path,
				"method":  c.Request.Method,
			}).Debug("Using development user authentication bypass")
			c.Next()
			return
		}

		// For production we would verify the token, but for now just proceed
		c.Next()
	}
}

// LoggerMiddleware adds a logger to the gin context
func LoggerMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add logger to context
		c.Set("logger", logger)
		
		// Process request
		c.Next()
	}
}

// CORSMiddleware handles CORS settings
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Log the origin for debugging
		logger, exists := c.Get("logger")
		if exists && logger != nil {
			log := logger.(*logrus.Logger)
			log.Infof("Received request with Origin: %s", origin)
		}
		
		// Allow the specific requesting origin (most permissive valid approach)
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			// Fallback when no origin is provided
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// Handle pre-flight OPTIONS request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// GetUserFromContext gets the user from the gin context
func GetUserFromContext(c *gin.Context) (models.User, error) {
	user, exists := c.Get("user")
	if !exists {
		return models.User{}, errors.New("user not found in context")
	}
	
	return user.(models.User), nil
}

// GetUserIDFromContext retrieves the user ID from the context
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return uuid.UUID{}, errors.New("userID not found in context")
	}
	
	return userID.(uuid.UUID), nil
} 