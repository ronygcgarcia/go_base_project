package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ronygcgarcia/go_base_project/config"
	"github.com/ronygcgarcia/go_base_project/migrations"
	"github.com/ronygcgarcia/go_base_project/routes"
)

func main() {
	if len(os.Args) > 1 {
		cmd := os.Args[1]

		switch cmd {
		case "migrate":
			config.ConnectDatabase()
			if err := migrations.Run(config.DB); err != nil {
				fmt.Println("❌ Migration error:", err)
				os.Exit(1)
			}
			fmt.Println("✅ Migration completed successfully.")
			return

		case "rollback":
			config.ConnectDatabase()
			if err := migrations.Rollback(config.DB); err != nil {
				fmt.Println("❌ Rollback error:", err)
				os.Exit(1)
			}
			fmt.Println("✅ Rollback completed successfully.")
			return

		case "make:migration":
			if len(os.Args) < 3 {
				fmt.Println("❌ Migration name is required.")
				os.Exit(1)
			}
			name := strings.ToLower(os.Args[2])
			if err := migrations.CreateMigrationFile(name); err != nil {
				fmt.Println("❌ Error creating migration:", err)
				os.Exit(1)
			}
			fmt.Println("✅ Migration file created successfully.")
			return
		}
	}

	// Run server
	r := routes.SetupRouter()
	host := os.Getenv("SERVER_HOST")
	if host == "" {
		host = "localhost" // fallback
	}

	env := strings.ToLower(os.Getenv("APP_ENV"))
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	if env == "production" {
		fmt.Printf("🔒 Starting production server at https://%s...\n", host)
		err := r.RunTLS(":443", "./certs/server.crt", "./certs/server.key")
		if err != nil {
			log.Fatalf("❌ Failed to start HTTPS server: %v", err)
		}
	} else {
		fmt.Printf("🚀 Starting development server at http://%s:%s...\n", host, port)
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("❌ Failed to start development server: %v", err)
		}
	}

}
