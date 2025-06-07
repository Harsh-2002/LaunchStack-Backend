package migrations

import (
	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	Name     string
	Migrate  func(db *gorm.DB) error
	Rollback func(db *gorm.DB) error
}

// AllMigrations returns a slice of all migrations
func AllMigrations() []Migration {
	return []Migration{
		// Add migrations in order
		{
			Name:     "CreateUsersTable",
			Migrate:  CreateUsersTable,
			Rollback: DropUsersTable,
		},
		{
			Name:     "CreateInstancesTable",
			Migrate:  CreateInstancesTable,
			Rollback: DropInstancesTable,
		},
		{
			Name:     "CreateResourceUsageTable",
			Migrate:  CreateResourceUsageTable,
			Rollback: func(db *gorm.DB) error { return db.Migrator().DropTable("resource_usages") },
		},
		// Add more migrations as needed
	}
} 