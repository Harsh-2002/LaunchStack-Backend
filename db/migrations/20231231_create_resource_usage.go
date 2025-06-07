package migrations

import (
	"github.com/launchstack/backend/models"
	"gorm.io/gorm"
)

// CreateResourceUsageTable creates the resource_usage table
func CreateResourceUsageTable(db *gorm.DB) error {
	// First check if the table exists
	if db.Migrator().HasTable(&models.ResourceUsage{}) {
		// Check if memory_limit column exists
		if !db.Migrator().HasColumn(&models.ResourceUsage{}, "memory_limit") {
			// Add the missing column
			if err := db.Exec("ALTER TABLE resource_usages ADD COLUMN memory_limit BIGINT").Error; err != nil {
				return err
			}
		}
		// Check if memory_percentage column exists
		if !db.Migrator().HasColumn(&models.ResourceUsage{}, "memory_percentage") {
			// Add the missing column
			if err := db.Exec("ALTER TABLE resource_usages ADD COLUMN memory_percentage FLOAT").Error; err != nil {
				return err
			}
		}
		
		// Check if timestamp columns exist
		if !db.Migrator().HasColumn(&models.ResourceUsage{}, "created_at") {
			if err := db.Exec("ALTER TABLE resource_usages ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()").Error; err != nil {
				return err
			}
		}
		if !db.Migrator().HasColumn(&models.ResourceUsage{}, "updated_at") {
			if err := db.Exec("ALTER TABLE resource_usages ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()").Error; err != nil {
				return err
			}
		}
		if !db.Migrator().HasColumn(&models.ResourceUsage{}, "deleted_at") {
			if err := db.Exec("ALTER TABLE resource_usages ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE").Error; err != nil {
				return err
			}
		}
		
		return nil
	}
	
	// Create the table from scratch if it doesn't exist
	return db.AutoMigrate(&models.ResourceUsage{})
} 