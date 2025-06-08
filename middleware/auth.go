package middleware

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/launchstack/backend/config"
	"github.com/launchstack/backend/db"
	"github.com/launchstack/backend/models"
	"github.com/MicahParks/keyfunc"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	jwksURL     string
	jwks        *keyfunc.JWKS
	jwksOnce    sync.Once
	jwksRefresh time.Duration = 12 * time.Hour
)

// initJWKS initializes the JWKS from Clerk
func initJWKS(clerkInstanceID string, logger *logrus.Logger) error {
	jwksURL = fmt.Sprintf("https://%s.clerk.accounts.dev/.well-known/jwks.json", clerkInstanceID)
	logger.Infof("Initializing JWKS from %s", jwksURL)
	
	options := keyfunc.Options{
		RefreshInterval: jwksRefresh,
		RefreshErrorHandler: func(err error) {
			logger.Errorf("Error refreshing JWKS: %v", err)
		},
	}
	
	var err error
	jwks, err = keyfunc.Get(jwksURL, options)
	if err != nil {
		logger.Errorf("Failed to get JWKS: %v", err)
		return err
	}
	
	logger.Info("JWKS initialized successfully")
	return nil
}

// AuthMiddleware validates the JWT token and adds the user to the context
func AuthMiddleware(clerkSecretKey string, logger *logrus.Logger, cfg *config.Config) gin.HandlerFunc {
	// Extract Clerk instance ID from the domain
	// The format is usually "something.clerk.accounts.dev"
	clerkInstanceID := strings.Split(cfg.Clerk.Issuer, ".")[0]
	
	// Initialize JWKS once
	jwksOnce.Do(func() {
		if err := initJWKS(clerkInstanceID, logger); err != nil {
			logger.Errorf("Failed to initialize JWKS: %v", err)
		}
	})
	
	return func(c *gin.Context) {
		// Skip authentication for public endpoints
		if isPublicEndpoint(c.Request.URL.Path) {
			logger.WithField("path", c.Request.URL.Path).Debug("Skipping authentication for public endpoint")
			c.Next()
			return
		}

		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("Invalid Authorization format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		// Get the token
		tokenString := parts[1]
		
		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the algorithm
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			
			// Special case for test tokens
			if kid, ok := token.Header["kid"].(string); ok && kid == "test-key-1" {
				// For test tokens, load the public key from file
				publicKeyBytes, err := os.ReadFile("test_public_key.pem")
				if err != nil {
					logger.WithError(err).Error("Failed to read test public key file")
					return nil, err
				}
				
				block, _ := pem.Decode(publicKeyBytes)
				if block == nil {
					logger.Error("Failed to parse PEM block containing the test public key")
					return nil, fmt.Errorf("failed to parse PEM block containing the public key")
				}
				
				pub, err := x509.ParsePKIXPublicKey(block.Bytes)
				if err != nil {
					logger.WithError(err).Error("Failed to parse test public key")
					return nil, err
				}
				
				rsaPublicKey, ok := pub.(*rsa.PublicKey)
				if !ok {
					logger.Error("Test key is not an RSA public key")
					return nil, fmt.Errorf("not an RSA public key")
				}
				
				logger.Info("Using test token authentication")
				return rsaPublicKey, nil
			}
			
			// Get the key from JWKS for normal tokens
			return jwks.Keyfunc(token)
		})
		
		if err != nil {
			logger.WithError(err).Error("Failed to parse token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		
		// Check if token is valid
		if !token.Valid {
			logger.Error("Token is invalid")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		
		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logger.Error("Could not extract claims from token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
		
		// Extract user ID from claims
		var clerkUserID string
		
		// First try to get from user_id claim (our custom claim)
		if userID, ok := claims["user_id"].(string); ok && userID != "" {
			clerkUserID = userID
		} else if sub, ok := claims["sub"].(string); ok && sub != "" {
			// Fall back to standard sub claim
			clerkUserID = sub
		} else {
			logger.Error("No user identifier found in token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: no user identifier"})
			c.Abort()
			return
		}
		
		// Log successful token validation
		logger.WithField("clerk_user_id", clerkUserID).Info("Token validated successfully")
		
		// Get user from database
		user, err := db.FindUserByClerkID(clerkUserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// User not found - could happen if they signed up but webhook hasn't processed yet
				logger.WithField("clerk_user_id", clerkUserID).Warn("User not found in database")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			} else {
				// Database error
				logger.WithError(err).Error("Database error when fetching user")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
			c.Abort()
			return
		}

		// Add user to context
		c.Set("userID", user.ID)
		c.Set("user", user)
		
		logger.WithFields(logrus.Fields{
			"user_id": user.ID.String(),
			"email":   user.Email,
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
		}).Debug("User authenticated successfully")
		
		c.Next()
	}
}

// isPublicEndpoint checks if an endpoint should skip authentication
func isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/api/v1/health",
		"/api/v1/health/",
		"/health",
		"/api/v1/auth/webhook",
		"/api/v1/auth/webhook/",
		"/api/v1/webhooks/clerk",
		"/api/v1/webhooks/clerk/",
		"/api/v1/webhooks/paypal",
		"/api/v1/webhooks/paypal/",
	}
	
	for _, publicPath := range publicPaths {
		if path == publicPath {
			return true
		}
	}
	
	return false
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