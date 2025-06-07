package main

import (
	"fmt"
	"os"
	
	"github.com/joho/godotenv"
	"github.com/launchstack/backend/config"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file, using environment variables")
	}
	
	// Print Docker host from environment
	dockerHost := os.Getenv("DOCKER_HOST")
	fmt.Printf("DOCKER_HOST from environment: %s\n", dockerHost)
	
	// Load configuration
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}
	
	// Print Docker host from config
	fmt.Printf("Docker.Host from config: %s\n", cfg.Docker.Host)
	
	// Print other Docker configuration
	fmt.Printf("Docker.Network: %s\n", cfg.Docker.Network)
	fmt.Printf("Docker.NetworkSubnet: %s\n", cfg.Docker.NetworkSubnet)
	fmt.Printf("Docker.N8NContainerPort: %d\n", cfg.Docker.N8NContainerPort)
	
	// Compare the values
	if dockerHost == cfg.Docker.Host {
		fmt.Println("✅ Docker host configuration is consistent")
	} else {
		fmt.Printf("❌ Inconsistent Docker host configuration: env=%s, config=%s\n", 
			dockerHost, cfg.Docker.Host)
	}
} 