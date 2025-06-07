package migrations

import (
	"github.com/launchstack/backend/models"
	"gorm.io/gorm"
)

// CreateUsersTable creates the users table
func CreateUsersTable(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{})
}

// DropUsersTable drops the users table
func DropUsersTable(db *gorm.DB) error {
	return db.Migrator().DropTable("users")
} 