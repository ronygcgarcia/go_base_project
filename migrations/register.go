// Package migrations
// This file exposes the public API for migration registration.
// Internally, it delegates to migrations/config/registry.go

package migrations

import (
	"github.com/ronygcgarcia/go_base_project/migrations/config"
)

// Migration is the public alias for config.Migration
type Migration = config.Migration

// Register registers a migration struct into the registry.
// This allows the migration to be executed via Run or Rollback.
func Register(m Migration) {
	config.Register(m)
}

// All returns all registered migrations in order.
func All() []Migration {
	return config.All()
}
