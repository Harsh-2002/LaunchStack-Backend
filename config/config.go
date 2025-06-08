package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server struct {
		Port         int
		Environment  string
		JWTSecret    string
		BackendURL   string
		FrontendURL  string
		Domain       string
	}
	Database struct {
		URL string
	}
	Clerk struct {
		SecretKey        string
		WebhookSecret    string
		PublishableKey   string
		Issuer           string
	}
	PayPal struct {
		DisablePayments  bool
		APIKey           string
		Secret           string
		Mode             string
	}
	Docker struct {
		Host            string
		Network         string
		NetworkSubnet   string
		N8NContainerPort int
	}
	N8N struct {
		BaseImage      string
		DataDir        string
		PortRangeStart int
		PortRangeEnd   int
		WebhookSecret  string
	}
	CORS struct {
		Origins []string
	}
	Monitoring struct {
		Interval time.Duration
		LogLevel string
	}
}

// NewConfig creates a new Config struct from environment variables
func NewConfig() (*Config, error) {
	config := &Config{}

	// Server configuration
	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}
	config.Server.Port = port
	config.Server.Environment = getEnv("APP_ENV", "development")
	config.Server.JWTSecret = getEnv("JWT_SECRET", "")
	if config.Server.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	config.Server.BackendURL = getEnv("BACKEND_URL", "http://localhost:8080")
	config.Server.FrontendURL = getEnv("FRONTEND_URL", "http://localhost:3000")
	config.Server.Domain = getEnv("DOMAIN", "launchstack.io")

	// Database configuration
	config.Database.URL = getEnv("DATABASE_URL", "")
	if config.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	// Clerk configuration
	config.Clerk.SecretKey = getEnv("CLERK_SECRET_KEY", "")
	if config.Clerk.SecretKey == "" {
		return nil, fmt.Errorf("CLERK_SECRET_KEY is required")
	}
	config.Clerk.WebhookSecret = getEnv("CLERK_WEBHOOK_SECRET", "")
	config.Clerk.PublishableKey = getEnv("NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY", "")
	config.Clerk.Issuer = getEnv("CLERK_ISSUER", "glad-starling-70.clerk.accounts.dev")

	// PayPal configuration
	disablePayments := getEnv("DISABLE_PAYMENTS", "false")
	config.PayPal.DisablePayments = disablePayments == "true"
	config.PayPal.APIKey = getEnv("PAYPAL_API_KEY", "")
	config.PayPal.Secret = getEnv("PAYPAL_SECRET", "")
	config.PayPal.Mode = getEnv("PAYPAL_MODE", "sandbox")

	// Docker configuration
	config.Docker.Host = getEnv("DOCKER_HOST", "http://10.1.1.81:2375")
	config.Docker.Network = getEnv("DOCKER_NETWORK", "n8n")
	config.Docker.NetworkSubnet = getEnv("DOCKER_NETWORK_SUBNET", "10.1.2.0/24")
	
	n8nContainerPort, err := strconv.Atoi(getEnv("N8N_CONTAINER_PORT", "5678"))
	if err != nil {
		return nil, fmt.Errorf("invalid N8N_CONTAINER_PORT: %w", err)
	}
	config.Docker.N8NContainerPort = n8nContainerPort

	// N8N configuration
	config.N8N.BaseImage = getEnv("N8N_BASE_IMAGE", "n8nio/n8n:latest")
	config.N8N.DataDir = getEnv("N8N_DATA_DIR", "/opt/n8n/data")
	config.N8N.WebhookSecret = getEnv("N8N_WEBHOOK_SECRET", "n8n_webhook_" + config.Server.JWTSecret[:8])
	portStart, err := strconv.Atoi(getEnv("N8N_PORT_RANGE_START", "5000"))
	if err != nil {
		return nil, fmt.Errorf("invalid N8N_PORT_RANGE_START: %w", err)
	}
	config.N8N.PortRangeStart = portStart
	
	portEnd, err := strconv.Atoi(getEnv("N8N_PORT_RANGE_END", "6000"))
	if err != nil {
		return nil, fmt.Errorf("invalid N8N_PORT_RANGE_END: %w", err)
	}
	config.N8N.PortRangeEnd = portEnd

	// CORS configuration
	corsOrigins := getEnv("CORS_ORIGINS", "*")
	if corsOrigins == "*" {
		config.CORS.Origins = []string{"*"}
	} else {
		config.CORS.Origins = strings.Split(corsOrigins, ",")
	}

	// Monitoring configuration
	monitorInterval, err := time.ParseDuration(getEnv("RESOURCE_MONITOR_INTERVAL", "30s"))
	if err != nil {
		return nil, fmt.Errorf("invalid RESOURCE_MONITOR_INTERVAL: %w", err)
	}
	config.Monitoring.Interval = monitorInterval
	config.Monitoring.LogLevel = getEnv("LOG_LEVEL", "info")

	return config, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
} 