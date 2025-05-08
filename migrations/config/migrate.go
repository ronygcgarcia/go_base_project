package config

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

func ensureSchemaTable(db *gorm.DB) {
	db.AutoMigrate(&SchemaMigration{})
}

func hasMigration(db *gorm.DB, name string) bool {
	var count int64
	db.Model(&SchemaMigration{}).Where("name = ?", name).Count(&count)
	return count > 0
}

func saveMigration(db *gorm.DB, name string) {
	db.Create(&SchemaMigration{Name: name, AppliedAt: time.Now()})
}

func removeMigration(db *gorm.DB, name string) {
	db.Where("name = ?", name).Delete(&SchemaMigration{})
}

func Run(db *gorm.DB) error {
	ensureSchemaTable(db)

	fmt.Println("Executing migrations...")
	for _, m := range All() {
		if hasMigration(db, m.Name()) {
			fmt.Printf("↷ %s Already applied, skipping\n", m.Name())
			continue
		}

		fmt.Printf("→ %s\n", m.Name())
		if err := m.Up(db); err != nil {
			return fmt.Errorf("Error while executing %s: %w", m.Name(), err)
		}

		saveMigration(db, m.Name())
	}

	return nil
}

func Rollback(db *gorm.DB) error {
	ensureSchemaTable(db)

	fmt.Println("Reverting migrations...")
	for i := len(All()) - 1; i >= 0; i-- {
		m := All()[i]
		if !hasMigration(db, m.Name()) {
			fmt.Printf("⤷ %s has not been applied, skipping\n", m.Name())
			continue
		}

		fmt.Printf("↩ Rolling back %s...\n", m.Name())
		if err := m.Down(db); err != nil {
			return fmt.Errorf("rollback failed for %s: %w", m.Name(), err)
		}

		removeMigration(db, m.Name())
	}

	return nil
}

func RollbackOne(db *gorm.DB) error {
	ensureSchemaTable(db)

	var last SchemaMigration
	result := db.Order("applied_at desc").Limit(1).Find(&last)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		fmt.Println("⤷ No migrations have been applied.")
		return nil
	}

	for _, m := range All() {
		if m.Name() == last.Name {
			fmt.Printf("↩ Rolling back migration: %s\n", m.Name())
			if err := m.Down(db); err != nil {
				return fmt.Errorf("error rolling back migration %s: %w", m.Name(), err)
			}
			db.Delete(&last)
			fmt.Println("✅ Migration rolled back.")
			return nil
		}
	}

	fmt.Printf("⚠️ Migration %s not found in registry.\n", last.Name)
	return nil
}
