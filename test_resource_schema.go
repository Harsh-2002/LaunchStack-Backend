package main

import (
	"fmt"
	"os"

	"github.com/launchstack/backend/db"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	// Use the same database connection as the main application
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		// Default connection string - adjust as needed
		dbURL = "postgresql://postgres:postgres@db:5432/launchstack?sslmode=disable"
	}
	
	logger.Infof("Using database connection: %s", dbURL)

	// Connect to the database
	logger.Info("Connecting to database...")
	if err := db.Initialize(dbURL); err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}

	// Check if memory_limit column exists
	var count int64
	query := `
		SELECT COUNT(*) 
		FROM information_schema.columns 
		WHERE table_name = 'resource_usages' 
		AND column_name = 'memory_limit'
	`
	
	result := db.DB.Raw(query).Scan(&count)
	if result.Error != nil {
		logger.WithError(result.Error).Fatal("Failed to check resource_usages schema")
	}

	if count > 0 {
		logger.Info("memory_limit column exists in resource_usages table")
	} else {
		logger.Error("memory_limit column does NOT exist in resource_usages table")
		
		// Try to add the column
		logger.Info("Attempting to add memory_limit column...")
		alterQuery := "ALTER TABLE resource_usages ADD COLUMN memory_limit BIGINT"
		if err := db.DB.Exec(alterQuery).Error; err != nil {
			logger.WithError(err).Fatal("Failed to add memory_limit column")
		} else {
			logger.Info("Successfully added memory_limit column")
		}
	}

	// Check if memory_percentage column exists
	query = `
		SELECT COUNT(*) 
		FROM information_schema.columns 
		WHERE table_name = 'resource_usages' 
		AND column_name = 'memory_percentage'
	`
	
	result = db.DB.Raw(query).Scan(&count)
	if result.Error != nil {
		logger.WithError(result.Error).Fatal("Failed to check resource_usages schema")
	}

	if count > 0 {
		logger.Info("memory_percentage column exists in resource_usages table")
	} else {
		logger.Error("memory_percentage column does NOT exist in resource_usages table")
		
		// Try to add the column
		logger.Info("Attempting to add memory_percentage column...")
		alterQuery := "ALTER TABLE resource_usages ADD COLUMN memory_percentage FLOAT"
		if err := db.DB.Exec(alterQuery).Error; err != nil {
			logger.WithError(err).Fatal("Failed to add memory_percentage column")
		} else {
			logger.Info("Successfully added memory_percentage column")
		}
	}

	// List all columns in resource_usages table
	type Column struct {
		ColumnName string
		DataType   string
	}
	
	var columns []Column
	
	listQuery := `
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'resource_usages' 
		ORDER BY column_name
	`
	
	result = db.DB.Raw(listQuery).Scan(&columns)
	if result.Error != nil {
		logger.WithError(result.Error).Fatal("Failed to list resource_usages columns")
	}
	
	logger.Info("Columns in resource_usages table:")
	for _, col := range columns {
		fmt.Printf("  %s (%s)\n", col.ColumnName, col.DataType)
	}
} 