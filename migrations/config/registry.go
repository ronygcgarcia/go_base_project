package config

import "gorm.io/gorm"

type Migration interface {
	Name() string
	Up(db *gorm.DB) error
	Down(db *gorm.DB) error
}

var registry []Migration

func Register(m Migration) {
	registry = append(registry, m)
}

func All() []Migration {
	return registry
}
