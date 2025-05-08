package main

import (
	"fmt"
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
			err := migrations.Run(config.DB)
			if err != nil {
				fmt.Println("Error while executing migrations:", err)
				os.Exit(1)
			}
			fmt.Println("✅ Migration finished.")
			return

		case "rollback":
			config.ConnectDatabase()
			err := migrations.Rollback(config.DB)
			if err != nil {
				fmt.Println("Error while rolling back:", err)
				os.Exit(1)
			}
			fmt.Println("✅ Rollback finished.")
			return

		case "make:migration":
			if len(os.Args) < 3 {
				fmt.Println("❌ You should provide a migration name.")
				os.Exit(1)
			}
			name := strings.ToLower(os.Args[2])
			err := migrations.CreateMigrationFile(name)
			if err != nil {
				fmt.Println("Error creating migration:", err)
				os.Exit(1)
			}
			fmt.Println("✅ Migration created.")
			return
		}
	}

	// Modo servidor
	r := routes.SetupRouter()
	r.Run(":8080")
}
