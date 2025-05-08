package config

import (
    "fmt"
    "log"
    "os"

    "github.com/joho/godotenv"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

)

var DB *gorm.DB

func ConnectDatabase() {
    // Carga las variables de entorno desde .env (solo en dev/local)
    err := godotenv.Load()
    if err != nil {
        log.Println("Advertencia: No se pudo cargar .env, usando variables del sistema")
    }

    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    dbname := os.Getenv("DB_NAME")
    sslmode := os.Getenv("DB_SSLMODE")

    dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        host, port, user, password, dbname, sslmode)

    database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Error al conectar la base de datos:", err)
    }

    // Migración automática (puedes modularizar esto después)
    // database.AutoMigrate(&models.User{})

    DB = database
}
