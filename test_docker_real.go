package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/launchstack/backend/config"
	"github.com/launchstack/backend/container"
	"github.com/launchstack/backend/models"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.Info("Starting Docker test...")

	// Load configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Docker client
	logger.Info("Initializing Docker client...")
	dockerClient, err := container.NewDockerClient("")
	if err != nil {
		log.Fatalf("Failed to initialize Docker client: %v", err)
	}
	logger.Info("Docker client initialized successfully")

	// Create container manager
	logger.Info("Creating container manager...")
	containerManager := container.NewManager(dockerClient, cfg, logger)
	logger.Info("Container manager created successfully")

	// Create a test user with unlimited resources
	testUser := models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Plan:      "unlimited", // Using unlimited plan to avoid resource limits
	}

	// Create a test instance
	instanceReq := models.Instance{
		Name:        "Test Instance",
		Description: "Test instance created by test_docker_real.go",
	}

	// Create the instance
	logger.Info("Creating test instance...")
	instance, err := containerManager.CreateInstance(context.Background(), testUser, instanceReq)
	if err != nil {
		log.Fatalf("Failed to create instance: %v", err)
	}
	logger.WithFields(logrus.Fields{
		"instance_id":   instance.ID,
		"instance_name": instance.Name,
		"status":        instance.Status,
		"url":           instance.URL,
	}).Info("Instance created successfully")

	// Wait for a while to see if the instance starts properly
	logger.Info("Waiting for 10 seconds...")
	time.Sleep(10 * time.Second)

	// Stop the instance
	logger.Info("Stopping the instance...")
	err = containerManager.StopInstance(context.Background(), instance.ID)
	if err != nil {
		logger.Warnf("Failed to stop instance: %v", err)
	} else {
		logger.Info("Instance stopped successfully")
	}

	// Wait for a while
	logger.Info("Waiting for 5 seconds...")
	time.Sleep(5 * time.Second)

	// Start the instance
	logger.Info("Starting the instance...")
	err = containerManager.StartInstance(context.Background(), instance.ID)
	if err != nil {
		logger.Warnf("Failed to start instance: %v", err)
	} else {
		logger.Info("Instance started successfully")
	}

	// Wait for a while
	logger.Info("Waiting for 5 seconds...")
	time.Sleep(5 * time.Second)

	// Cleanup - delete the instance
	logger.Info("Deleting the instance...")
	err = containerManager.DeleteInstance(context.Background(), instance.ID)
	if err != nil {
		logger.Warnf("Failed to delete instance: %v", err)
	} else {
		logger.Info("Instance deleted successfully")
	}

	logger.Info("Test completed successfully")
	fmt.Println("Test completed. Please check the logs for details.")
} 