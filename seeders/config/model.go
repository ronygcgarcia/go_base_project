package config

import "time"

type SchemaSeeder struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"uniqueIndex"`
	AppliedAt time.Time
}
