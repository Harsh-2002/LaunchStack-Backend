package db

import (
	"fmt"
	"time"

	"github.com/launchstack/backend/models"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database connection
var DB *gorm.DB

// MigrationRecord tracks when migrations have been run
type MigrationRecord struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"uniqueIndex"`
	AppliedAt time.Time
}

// Initialize sets up the database connection
func Initialize(dsn string) error {
	var err error
	
	// Configure GORM to be silent in production
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}
	
	// Connect to the database
	DB, err = gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	return nil
}

// RunMigrations runs database migrations only if they haven't been run before
func RunMigrations() error {
	// First, create the migrations table if it doesn't exist
	err := DB.AutoMigrate(&MigrationRecord{})
	if err != nil {
		return fmt.Errorf("failed to create migration tracking table: %w", err)
	}
	
	// Check if we've run migrations before
	var count int64
	DB.Model(&MigrationRecord{}).Count(&count)
	
	// If we already have migration records, check if we need to run again
	if count > 0 {
		// Get the latest migration record
		var lastMigration MigrationRecord
		if err := DB.Order("applied_at DESC").First(&lastMigration).Error; err == nil {
			// If we've run migrations in the last 24 hours, skip
			if time.Since(lastMigration.AppliedAt) < 24*time.Hour {
				return nil
			}
		}
	}
	
	// Run migrations and record that we did
	err = DB.AutoMigrate(
		&models.User{},
		&models.Instance{},
		&models.ResourceUsage{},
		&models.Payment{},
	)
	
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	
	// Record that we ran migrations
	migrationRecord := MigrationRecord{
		Name:      fmt.Sprintf("migration-%s", time.Now().Format("20060102-150405")),
		AppliedAt: time.Now(),
	}
	
	if err := DB.Create(&migrationRecord).Error; err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}
	
	return nil
}

// RunMigrationsWithLogger runs migrations with logging
func RunMigrationsWithLogger(logger *logrus.Logger) error {
	// Check if migrations have been run recently
	var count int64
	DB.Model(&MigrationRecord{}).Count(&count)
	
	if count > 0 {
		// Get the latest migration record
		var lastMigration MigrationRecord
		if err := DB.Order("applied_at DESC").First(&lastMigration).Error; err == nil {
			timeSince := time.Since(lastMigration.AppliedAt)
			if timeSince < 24*time.Hour {
				logger.Infof("Skipping migrations - last run %s ago", timeSince.Round(time.Second))
				return nil
			}
			logger.Infof("Running migrations - last run %s ago", timeSince.Round(time.Second))
		}
	} else {
		logger.Info("Running migrations for the first time")
	}
	
	// Run the migrations
	return RunMigrations()
}

// Close closes the database connection
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
} 