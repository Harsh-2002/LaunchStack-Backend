package migrations

import (
	"github.com/launchstack/backend/models"
	"gorm.io/gorm"
)

// CreateInstancesTable creates the instances table
func CreateInstancesTable(db *gorm.DB) error {
	return db.AutoMigrate(&models.Instance{})
}

// DropInstancesTable drops the instances table
func DropInstancesTable(db *gorm.DB) error {
	return db.Migrator().DropTable("instances")
} 