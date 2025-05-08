package config

import "gorm.io/gorm"

type Seeder interface {
	Name() string
	Up(db *gorm.DB) error
	Down(db *gorm.DB) error
}

var registry []Seeder

func Register(s Seeder) {
	registry = append(registry, s)
}

func All() []Seeder {
	return registry
}
