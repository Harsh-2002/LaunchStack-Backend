package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/launchstack/backend/db/migrations"
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

// Initialize database connection
func InitDB() error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get database connection details from environment variables
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "npg_eiCzc53PmMRS")
	dbHost := getEnv("DB_HOST", "10.1.1.82")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "launchstack")
	
	// Configure database connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbHost, dbUser, dbPassword, dbName, dbPort)
	
	// Configure GORM logger
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
	
	// Connect to database
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	log.Println("Connected to TimescaleDB successfully")
	
	// Auto migrate schemas
	err = migrateSchemas()
	if err != nil {
		return fmt.Errorf("failed to migrate database schemas: %w", err)
	}
	
	log.Println("Database migrations completed successfully")
	return nil
}

// Get environment variable with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// Migrate database schemas
func migrateSchemas() error {
	// Auto migrate models
	// This creates/updates tables based on your struct definitions
	// Add all models that need to be migrated here
	err := DB.AutoMigrate(
		&models.User{},
		&models.Instance{},
		&models.ResourceUsage{},
		// Add other models as needed
	)
	
	if err != nil {
		return err
	}
	
	// Create hypertable for ResourceUsage if it doesn't exist
	// This needs to be done after the table is created by GORM
	DB.Exec("SELECT create_hypertable('resource_usages', 'timestamp', if_not_exists => TRUE)")
	
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
				// Always run resource usage migration to ensure the column exists
				if err := migrations.CreateResourceUsageTable(DB); err != nil {
					return fmt.Errorf("failed to run resource usage migration: %w", err)
				}
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
	
	// Explicitly run resource usage migration to ensure the column exists
	if err := migrations.CreateResourceUsageTable(DB); err != nil {
		return fmt.Errorf("failed to run resource usage migration: %w", err)
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
				
				// Always run our custom migrations
				logger.Info("Running custom migrations...")
				// Ensure resource_usages table has memory_limit column
				if err := migrations.CreateResourceUsageTable(DB); err != nil {
					logger.Warnf("Failed to run resource usage migration: %v", err)
				} else {
					logger.Info("Resource usage migration completed successfully")
				}
				
				if err := RunIPAddressMigration(); err != nil {
					logger.Warnf("Failed to run IP address migration: %v", err)
				} else {
					logger.Info("IP address migration completed successfully")
				}
				return nil
			}
			logger.Infof("Running migrations - last run %s ago", timeSince.Round(time.Second))
		}
	} else {
		logger.Info("Running migrations for the first time")
	}
	
	// Run the migrations
	err := RunMigrations()
	if err != nil {
		return err
	}
	
	// Run our custom migrations
	logger.Info("Running custom migrations...")
	if err := RunIPAddressMigration(); err != nil {
		logger.Warnf("Failed to run IP address migration: %v", err)
	} else {
		logger.Info("IP address migration completed successfully")
	}
	
	return nil
}

// Close closes the database connection
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
} 