package migrations

import (
    "fmt"
    "gorm.io/gorm"
    "time"
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

    fmt.Println("Ejecutando migraciones...")
    for _, m := range All() {
        if hasMigration(db, m.Name()) {
            fmt.Printf("↷ %s ya fue aplicada, saltando\n", m.Name())
            continue
        }

        fmt.Printf("→ %s\n", m.Name())
        if err := m.Up(db); err != nil {
            return fmt.Errorf("error en %s: %w", m.Name(), err)
        }

        saveMigration(db, m.Name())
    }

    return nil
}

func Rollback(db *gorm.DB) error {
    ensureSchemaTable(db)

    fmt.Println("Revirtiendo migraciones...")
    for i := len(All()) - 1; i >= 0; i-- {
        m := All()[i]
        if !hasMigration(db, m.Name()) {
            fmt.Printf("⤷ %s no ha sido aplicada, saltando\n", m.Name())
            continue
        }

        fmt.Printf("↩ %s\n", m.Name())
        if err := m.Down(db); err != nil {
            return fmt.Errorf("error en rollback %s: %w", m.Name(), err)
        }

        removeMigration(db, m.Name())
    }

    return nil
}
