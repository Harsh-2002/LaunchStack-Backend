package main

import (
	"fmt"
	
	"github.com/joho/godotenv"
	"github.com/launchstack/backend/config"
	"github.com/launchstack/backend/db"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		logger.Warning("Error loading .env file, using environment variables")
	}
	
	// Initialize configuration
	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Initialize database connection
	logger.Info("Connecting to database...")
	err = db.Initialize(cfg.Database.URL)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	logger.Info("Successfully connected to database")
	
	// List all tables in the database
	var tables []string
	err = db.DB.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tables).Error
	if err != nil {
		logger.Fatalf("Failed to query tables: %v", err)
	}
	
	// Print the tables
	fmt.Println("Tables in the database:")
	for i, table := range tables {
		fmt.Printf("%d. %s\n", i+1, table)
		
		// Count rows in the table
		var count int64
		err = db.DB.Table(table).Count(&count).Error
		if err != nil {
			logger.Warnf("Failed to count rows in table %s: %v", table, err)
			continue
		}
		fmt.Printf("   - Rows: %d\n", count)
		
		// Get columns for the table
		type Column struct {
			ColumnName string
			DataType   string
		}
		var columns []Column
		err = db.DB.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_schema = 'public' AND table_name = ?", table).Scan(&columns).Error
		if err != nil {
			logger.Warnf("Failed to get columns for table %s: %v", table, err)
			continue
		}
		
		fmt.Println("   - Columns:")
		for _, col := range columns {
			fmt.Printf("     - %s (%s)\n", col.ColumnName, col.DataType)
		}
		fmt.Println()
	}
} 