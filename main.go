package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ronygcgarcia/go_base_project/commands"
	"github.com/ronygcgarcia/go_base_project/routes"
)

func main() {
	// Handle CLI commands first
	if commands.HandleCLI() {
		return
	}

	// Run server
	r := routes.SetupRouter()

	host := os.Getenv("SERVER_HOST")
	if host == "" {
		host = "localhost"
	}

	env := strings.ToLower(os.Getenv("APP_ENV"))
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	crtPath := os.Getenv("SSL_CERT_PATH")
	if crtPath == "" {
		crtPath = "./certs/server.crt"
	}
	keyPath := os.Getenv("SSL_KEY_PATH")
	if keyPath == "" {
		keyPath = "./certs/server.key"
	}

	fmt.Printf("🧭 Environment: %s | Host: %s\n", env, host)

	if env == "production" {
		fmt.Printf("🔒 Starting production server at https://%s...\n", host)
		err := r.RunTLS(":443", crtPath, keyPath)
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
