package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/launchstack/backend/container"
)

func main() {
	// Test the generateContainerName and generateEasySubdomain functions
	
	// Sample test cases
	testInstances := []struct {
		Name string
		ID   string
	}{
		{"My First Instance", "f47ac10b-58cc-4372-a567-0e02b2c3d479"},
		{"Development Server", "550e8400-e29b-41d4-a716-446655440000"},
		{"Production App", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		{"Test Environment", "7c9e6679-7425-40de-944b-e07fc1f90ae7"},
		{"Analytics Dashboard", "a8098c1a-f86e-11da-bd1a-00112444be1e"},
	}
	
	fmt.Println("Testing subdomain generation:")
	fmt.Println("=============================")
	
	for _, test := range testInstances {
		userID, _ := uuid.Parse(test.ID)
		
		// Use the container package's functions
		containerName := container.GenerateContainerName(userID, test.Name)
		subdomain := container.GenerateEasySubdomain(containerName)
		
		fmt.Printf("Instance: '%s'\n", test.Name)
		fmt.Printf("Container name: %s\n", containerName)
		fmt.Printf("Subdomain: %s\n", subdomain)
		fmt.Println("-----------------------------")
	}
} 