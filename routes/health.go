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

// Current API version
const ApiVersion = "1.0.0"

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      string    `json:"status"`
	Version     string    `json:"version"`
	Environment string    `json:"environment"`
	Timestamp   time.Time `json:"timestamp"`
	Database    struct {
		Status       string        `json:"status"`
		ResponseTime time.Duration `json:"response_time_ms"`
	} `json:"database"`
	System struct {
		MemoryUsage float64 `json:"memory_usage_mb"`
		CPUCores    int     `json:"cpu_cores"`
		GoRoutines  int     `json:"go_routines"`
		Uptime      string  `json:"uptime"`
	} `json:"system"`
	ResponseTime time.Duration `json:"response_time_ms"`
}

// HealthCheck is a simple endpoint to verify the API is running
func HealthCheck(cfg *config.Config) gin.HandlerFunc {
	startTime := time.Now()
	return func(c *gin.Context) {
		// Calculate uptime
		uptime := time.Since(startTime).Round(time.Second).String()
		
		response := HealthResponse{
			Status:      "ok",
			Version:     ApiVersion,
			Environment: cfg.Server.Environment,
			Timestamp:   time.Now(),
		}

		// Check database connection with timing
		dbStartTime := time.Now()
		if err := db.DB.Exec("SELECT 1").Error; err != nil {
			response.Database.Status = "error"
			response.Status = "degraded"
		} else {
			response.Database.Status = "ok"
		}
		response.Database.ResponseTime = time.Since(dbStartTime).Round(time.Millisecond)

		// Get system metrics
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		
		response.System.MemoryUsage = float64(memStats.Alloc) / 1024 / 1024 // Convert to MB
		response.System.CPUCores = runtime.NumCPU()
		response.System.GoRoutines = runtime.NumGoroutine()
		response.System.Uptime = uptime

		// Calculate total response time
		response.ResponseTime = time.Since(time.Now().Add(-time.Millisecond)).Round(time.Millisecond)

		statusCode := http.StatusOK
		if response.Status != "ok" {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, response)
	}
}

// HealthCheckHandler handles health check requests with detailed information
func HealthCheckHandler(cfg *config.Config, logger *logrus.Logger) gin.HandlerFunc {
	startTime := time.Now()
	return func(c *gin.Context) {
		requestStartTime := time.Now()
		
		// Calculate uptime
		uptime := time.Since(startTime).Round(time.Second).String()
		
		response := HealthResponse{
			Status:      "ok",
			Version:     ApiVersion,
			Environment: cfg.Server.Environment,
			Timestamp:   time.Now(),
		}

		// Check database connection with timing
		dbStartTime := time.Now()
		if err := db.DB.Exec("SELECT 1").Error; err != nil {
			response.Database.Status = "error"
			response.Status = "degraded"
			logger.Errorf("Health check - database connection failed: %v", err)
		} else {
			response.Database.Status = "ok"
		}
		response.Database.ResponseTime = time.Since(dbStartTime).Round(time.Millisecond)

		// Get system metrics
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		
		response.System.MemoryUsage = float64(memStats.Alloc) / 1024 / 1024 // Convert to MB
		response.System.CPUCores = runtime.NumCPU()
		response.System.GoRoutines = runtime.NumGoroutine()
		response.System.Uptime = uptime

		// Calculate total response time
		response.ResponseTime = time.Since(requestStartTime).Round(time.Millisecond)

		statusCode := http.StatusOK
		if response.Status != "ok" {
			statusCode = http.StatusServiceUnavailable
		}
		
		logger.Infof("Health check executed: status=%s, response_time=%s", response.Status, response.ResponseTime)
		c.JSON(statusCode, response)
	}
}