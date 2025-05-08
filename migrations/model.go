package migrations

import "time"

type SchemaMigration struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"uniqueIndex"`
	AppliedAt time.Time
}
