// Package migrations
// This file exposes the public API for managing migrations via CLI or other tools.
// It delegates to migrations/config/migrate.go

package migrations

import (
	"github.com/ronygcgarcia/go_base_project/migrations/config"
	"gorm.io/gorm"
)

// Run executes all pending migrations.
func Run(db *gorm.DB) error {
	return config.Run(db)
}

// Rollback reverts all applied migrations in reverse order.
func Rollback(db *gorm.DB) error {
	return config.Rollback(db)
}

// RollbackOne reverts only the most recent migration.
func RollbackOne(db *gorm.DB) error {
	return config.RollbackOne(db)
}

// CreateMigrationFile generates a new migration file with boilerplate content.
func CreateMigrationFile(name string) error {
	return config.CreateMigrationFile(name)
}
