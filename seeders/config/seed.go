package config

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

func ensureSeederTable(db *gorm.DB) {
	db.AutoMigrate(&SchemaSeeder{})
}

func hasSeeder(db *gorm.DB, name string) bool {
	var count int64
	db.Model(&SchemaSeeder{}).Where("name = ?", name).Count(&count)
	return count > 0
}

func saveSeeder(db *gorm.DB, name string) {
	db.Create(&SchemaSeeder{Name: name, AppliedAt: time.Now()})
}

func removeSeeder(db *gorm.DB, name string) {
	db.Where("name = ?", name).Delete(&SchemaSeeder{})
}

func Run(db *gorm.DB) error {
	ensureSeederTable(db)

	fmt.Println("Seeding database...")
	for _, s := range All() {
		if hasSeeder(db, s.Name()) {
			fmt.Printf("↷ %s already applied, skipping\n", s.Name())
			continue
		}

		fmt.Printf("→ Applying seeder: %s\n", s.Name())
		if err := s.Up(db); err != nil {
			return fmt.Errorf("error in seeder %s: %w", s.Name(), err)
		}
		saveSeeder(db, s.Name())
	}
	return nil
}

func Rollback(db *gorm.DB) error {
	ensureSeederTable(db)

	fmt.Println("Rolling back all seeders...")
	for i := len(All()) - 1; i >= 0; i-- {
		s := All()[i]
		if !hasSeeder(db, s.Name()) {
			fmt.Printf("⤷ %s not applied, skipping\n", s.Name())
			continue
		}

		fmt.Printf("↩ Rolling back seeder: %s\n", s.Name())
		if err := s.Down(db); err != nil {
			return fmt.Errorf("error in rollback of %s: %w", s.Name(), err)
		}
		removeSeeder(db, s.Name())
	}
	return nil
}

func RollbackOne(db *gorm.DB) error {
	ensureSeederTable(db)

	var last SchemaSeeder
	result := db.Order("applied_at desc").Limit(1).Find(&last)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		fmt.Println("⤷ No seeders have been applied.")
		return nil
	}

	for _, s := range All() {
		if s.Name() == last.Name {
			fmt.Printf("↩ Rolling back seeder: %s\n", s.Name())
			if err := s.Down(db); err != nil {
				return fmt.Errorf("error rolling back seeder %s: %w", s.Name(), err)
			}
			removeSeeder(db, s.Name())
			fmt.Println("✅ Seeder rolled back.")
			return nil
		}
	}

	fmt.Printf("⚠️ Seeder %s not found in registry.\n", last.Name)
	return nil
}
