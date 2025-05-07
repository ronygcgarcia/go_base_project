package config

import (
    "log"
    "gorm.io/gorm"
    "gorm.io/driver/postgres"

    "github.com/ronygcgarcia/go_base_project/models"
)

var DB *gorm.DB

func ConnectDatabase() {
    dsn := "host=localhost user=postgres password=tuclave dbname=miapi port=5432 sslmode=disable"
    database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Error al conectar la base de datos:", err)
    }

    database.AutoMigrate(&models.User{})

    DB = database
}
