package main

import (
    "github.com/ronygcgarcia/go_base_project/routes"
)

func main() {
    r := routes.SetupRouter() // Llama a tu enrutador modular
    r.Run(":8080")            // Levanta el servidor en localhost:8080
}
