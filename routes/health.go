package routes

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/launchstack/backend/config"
	"github.com/launchstack/backend/db"
	"github.com/sirupsen/logrus"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      string    `json:"status"`
	Version     string    `json:"version"`
	Environment string    `json:"environment"`
	GoVersion   string    `json:"go_version"`
	Timestamp   time.Time `json:"timestamp"`
	Database    struct {
		Status  string `json:"status"`
		Message string `json:"message,omitempty"`
	} `json:"database"`
	Docker struct {
		Status  string `json:"status"`
		Message string `json:"message,omitempty"`
	} `json:"docker"`
	API struct {
		Endpoints []string `json:"endpoints"`
	} `json:"api"`
}

// HealthCheck is a simple endpoint to verify the API is running
func HealthCheck(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := HealthResponse{
			Status:      "ok",
			Version:     "0.1.0", // TODO: Get from build info
			Environment: cfg.Server.Environment,
			GoVersion:   runtime.Version(),
			Timestamp:   time.Now(),
		}

		// Check database connection
		if err := db.DB.Exec("SELECT 1").Error; err != nil {
			response.Database.Status = "error"
			response.Database.Message = "Database connection failed"
			response.Status = "degraded"
		} else {
			response.Database.Status = "ok"
		}

		// Check Docker connection
		// We assume Docker is working if the service is running
		// In a more complete implementation, you would perform a Docker API check here
		response.Docker.Status = "ok"

		statusCode := http.StatusOK
		if response.Status != "ok" {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, response)
	}
}

// HealthCheckHandler handles health check requests with detailed information
func HealthCheckHandler(cfg *config.Config, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := HealthResponse{
			Status:      "ok",
			Version:     "0.1.0", // TODO: Get from build info
			Environment: cfg.Server.Environment,
			GoVersion:   runtime.Version(),
			Timestamp:   time.Now(),
		}

		// Check database connection
		if err := db.DB.Exec("SELECT 1").Error; err != nil {
			response.Database.Status = "error"
			response.Database.Message = "Database connection failed"
			response.Status = "degraded"
			logger.Errorf("Health check - database connection failed: %v", err)
		} else {
			response.Database.Status = "ok"
		}

		// Check Docker connection
		// We assume Docker is working if the service is running
		response.Docker.Status = "ok"
		
		// List core API endpoints
		response.API.Endpoints = []string{
			"/api/instances",
			"/api/v1/instances",
			"/api/users/me",
			"/api/v1/users/me",
			"/api/auth/webhook",
			"/api/v1/auth/webhook",
			"/health",
			"/api/v1/health",
		}

		statusCode := http.StatusOK
		if response.Status != "ok" {
			statusCode = http.StatusServiceUnavailable
		}
		
		logger.Infof("Health check executed: status=%s", response.Status)
		c.JSON(statusCode, response)
	}
} 