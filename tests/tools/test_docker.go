package main

import (
	"context"
	"fmt"
	"os"
	
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file, using environment variables")
	}
	
	// Explicitly set Docker API endpoint
	dockerHost := "http://10.1.1.81:2375"
	os.Setenv("DOCKER_HOST", dockerHost)
	
	fmt.Printf("Connecting to Docker API endpoint: %s\n", dockerHost)
	
	// Create Docker client
	cli, err := client.NewClientWithOpts(
		client.WithHost(dockerHost),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		fmt.Printf("Error creating Docker client: %v\n", err)
		os.Exit(1)
	}
	
	// Test connection by listing containers
	fmt.Println("Listing containers...")
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		fmt.Printf("Error listing containers: %v\n", err)
		os.Exit(1)
	}
	
	// Display container information
	fmt.Printf("Found %d containers:\n", len(containers))
	for i, container := range containers {
		fmt.Printf("%d. ID: %.12s, Image: %s, State: %s, Names: %v\n", 
			i+1, container.ID, container.Image, container.State, container.Names)
	}
	
	// Test Docker version
	fmt.Println("\nChecking Docker version...")
	version, err := cli.ServerVersion(context.Background())
	if err != nil {
		fmt.Printf("Error getting Docker version: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Docker version: %s, API version: %s\n", version.Version, version.APIVersion)
	fmt.Println("Docker API connection successful!")
} 