// Package seeders
// This file exposes the public API for managing seeders via CLI or other tools.
// It delegates to seeders/config/seed.go

package seeders

import (
	"github.com/ronygcgarcia/go_base_project/seeders/config"
	"gorm.io/gorm"
)

// Run executes all seeders that have not yet been applied.
func Run(db *gorm.DB) error {
	return config.Run(db)
}

// Rollback reverts all applied seeders in reverse order.
func Rollback(db *gorm.DB) error {
	return config.Rollback(db)
}

// RollbackOne reverts only the most recently applied seeder.
func RollbackOne(db *gorm.DB) error {
	return config.RollbackOne(db)
}

// CreateSeederFile generates a new seeder file with boilerplate content.
func CreateSeederFile(name string) error {
	return config.CreateSeederFile(name)
}
