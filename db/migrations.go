package db

import (
	"fmt"
	"time"
)

// RunIPAddressMigration adds the ip_address column to the instances table
func RunIPAddressMigration() error {
	// Read the migration file from the database
	var migrationRecord MigrationRecord
	result := DB.Where("name = ?", "add_ip_address_column").First(&migrationRecord)
	
	// If migration already exists, skip it
	if result.Error == nil {
		return nil
	}
	
	// Run the migration
	migrationSQL := "ALTER TABLE instances ADD COLUMN IF NOT EXISTS ip_address VARCHAR(50);"
	if err := DB.Exec(migrationSQL).Error; err != nil {
		return fmt.Errorf("failed to add ip_address column: %w", err)
	}
	
	// Record the migration
	migrationRecord = MigrationRecord{
		Name:      "add_ip_address_column",
		AppliedAt: time.Now(),
	}
	
	if err := DB.Create(&migrationRecord).Error; err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}
	
	return nil
} 