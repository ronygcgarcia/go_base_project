package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/ronygcgarcia/go_base_project/config"
	"github.com/ronygcgarcia/go_base_project/migrations"
	"github.com/ronygcgarcia/go_base_project/seeders"
)

func HandleCLI() bool {
	if len(os.Args) < 2 {
		return false
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

	case "rollback:step":
		config.ConnectDatabase()
		if err := migrations.RollbackOne(config.DB); err != nil {
			fmt.Println("❌ Step rollback error:", err)
			os.Exit(1)
		}
		fmt.Println("✅ One migration rolled back.")
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

	case "seed":
		config.ConnectDatabase()
		if err := seeders.Run(config.DB); err != nil {
			fmt.Println("❌ Seeder error:", err)
			os.Exit(1)
		}
		fmt.Println("✅ Seeders applied successfully.")
		return true

	case "seed:rollback":
		config.ConnectDatabase()
		if err := seeders.Rollback(config.DB); err != nil {
			fmt.Println("❌ Seeder rollback error:", err)
			os.Exit(1)
		}
		fmt.Println("✅ Last seeder rolled back.")
		return true

	case "make:seeder":
		if len(os.Args) < 3 {
			fmt.Println("❌ Seeder name is required.")
			os.Exit(1)
		}
		name := strings.ToLower(os.Args[2])
		if err := seeders.CreateSeederFile(name); err != nil {
			fmt.Println("❌ Error creating seeder:", err)
			os.Exit(1)
		}
		fmt.Println("✅ Seeder file created successfully.")
		return true
	case "seed:rollback:step":
		config.ConnectDatabase()
		if err := seeders.RollbackOne(config.DB); err != nil {
			fmt.Println("❌ Seeder rollback (step) error:", err)
			os.Exit(1)
		}
		fmt.Println("✅ Last seeder rolled back.")
		return true

	}

	return false
}
