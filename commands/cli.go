package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/ronygcgarcia/go_base_project/config"
	"github.com/ronygcgarcia/go_base_project/migrations"
)

func HandleCLI() bool {
	if len(os.Args) < 2 {
		return false // no command given, continue to run server
	}

	cmd := os.Args[1]

	switch cmd {
	case "migrate":
		config.ConnectDatabase()
		if err := migrations.Run(config.DB); err != nil {
			fmt.Println("❌ Migration error:", err)
			os.Exit(1)
		}
		fmt.Println("✅ Migration completed successfully.")
		return true

	case "rollback":
		config.ConnectDatabase()
		if err := migrations.Rollback(config.DB); err != nil {
			fmt.Println("❌ Rollback error:", err)
			os.Exit(1)
		}
		fmt.Println("✅ Rollback completed successfully.")
		return true

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
		return true
	}

	return false // unknown command
}
