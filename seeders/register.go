// Package seeders
// This file exposes the public API for seeder registration.
// Internally, it delegates to seeders/config/registry.go

package seeders

import (
	"github.com/ronygcgarcia/go_base_project/seeders/config"
)

// Seeder is the public alias for config.Seeder
type Seeder = config.Seeder

// Register registers a seeder struct into the registry.
func Register(s Seeder) {
	config.Register(s)
}

// All returns all registered seeders in order.
func All() []Seeder {
	return config.All()
}
