package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/launchstack/backend/models"
)

func main() {
	// Parse command line flags
	formatFlag := flag.String("format", "text", "Output format (text or json)")
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Connect to database
	dbLogger := logger.New(
		log.New(os.Stderr, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get all instances that are not deleted
	var instances []models.Instance
	if err := db.Where("status != ?", models.StatusDeleted).Find(&instances).Error; err != nil {
		log.Fatalf("Failed to get instances: %v", err)
	}

	// Output in the requested format
	if *formatFlag == "json" {
		outputJSON(instances)
	} else {
		outputText(instances)
	}
}

func outputText(instances []models.Instance) {
	fmt.Printf("Found %d active instances:\n", len(instances))
	for _, instance := range instances {
		fmt.Printf("ID: %s, Name: %s, Host: %s, Status: %s\n",
			instance.ID.String(),
			instance.Name,
			instance.Host,
			instance.Status,
		)
	}
}

func outputJSON(instances []models.Instance) {
	// Create a list of simplified instances for JSON output
	type SimpleInstance struct {
		ID         string `json:"ID"`
		Name       string `json:"Name"`
		Host       string `json:"Host"`
		Status     string `json:"Status"`
		URL        string `json:"URL"`
		ContainerID string `json:"ContainerID"`
	}

	var simpleInstances []SimpleInstance
	for _, instance := range instances {
		simpleInstances = append(simpleInstances, SimpleInstance{
			ID:         instance.ID.String(),
			Name:       instance.Name,
			Host:       instance.Host,
			Status:     string(instance.Status),
			URL:        instance.URL,
			ContainerID: instance.ContainerID,
		})
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(simpleInstances, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal instances to JSON: %v", err)
	}

	fmt.Println(string(jsonData))
} 