package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func ActivateAuthFlow(flow string) error {
	var routeFile, registerFunc, routeTemplate string

	switch flow {
	case "client_credentials":
		routeFile = "routes/auth_client_credentials.go"
		registerFunc = "RegisterClientCredentialsAuth(r)"
		routeTemplate = getClientCredentialsTemplate()
	case "client_password":
		routeFile = "routes/auth_client_password.go"
		registerFunc = "RegisterClientPasswordAuth(r)"
		routeTemplate = getClientPasswordTemplate()
	default:
		return fmt.Errorf("unsupported auth type: %s", flow)
	}

	// Create auth route file if not exists
	if _, err := os.Stat(routeFile); os.IsNotExist(err) {
		if err := os.WriteFile(routeFile, []byte(routeTemplate), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", routeFile, err)
		}
		fmt.Println("✅ Created:", routeFile)
	} else {
		fmt.Println("ℹ️ Route already exists:", routeFile)
	}

	// Register route in api.go if not already registered
	apiFile := "routes/api.go"
	apiContent, err := os.ReadFile(apiFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", apiFile, err)
	}

	if !strings.Contains(string(apiContent), registerFunc) {
		newContent := strings.Replace(string(apiContent),
			"r := gin.Default()",
			"r := gin.Default()\n\t"+registerFunc,
			1,
		)
		if err := os.WriteFile(apiFile, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to update %s: %w", apiFile, err)
		}
		fmt.Println("✅ Registered auth route in:", apiFile)
	} else {
		fmt.Println("ℹ️ Auth route already registered in:", apiFile)
	}

	// Generate RSA key pair if not exists
	if _, err := os.Stat("keys/private.pem"); os.IsNotExist(err) {
		fmt.Println("🔐 Generating RSA key pair...")
		os.MkdirAll("keys", os.ModePerm)
		if err := generateKeys(); err != nil {
			return fmt.Errorf("failed to generate keys: %w", err)
		}
		fmt.Println("✅ Keys generated in ./keys/")
	} else {
		fmt.Println("ℹ️ Keys already exist in ./keys/")
	}

	// Generate corresponding migration
	if err := createAuthMigration(flow); err != nil {
		return fmt.Errorf("failed to create migration for auth type: %w", err)
	}

	return nil
}

func getClientCredentialsTemplate() string {
	return `package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ronygcgarcia/go_base_project/auth"
)

func RegisterClientCredentialsAuth(r *gin.Engine) {
	authGroup := r.Group("/auth")
	authGroup.POST("/token", func(c *gin.Context) {
		var body struct {
			ClientID     string ` + "`json:\"client_id\"`" + `
			ClientSecret string ` + "`json:\"client_secret\"`" + `
		}

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, gin.H{"error": "Invalid payload"})
			return
		}

		if body.ClientID != "my-client" || body.ClientSecret != "my-secret" {
			c.JSON(401, gin.H{"error": "Invalid client credentials"})
			return
		}

		token, err := auth.IssueClientToken(body.ClientID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(200, gin.H{"access_token": token, "token_type": "Bearer"})
	})
}`
}

func getClientPasswordTemplate() string {
	return `package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ronygcgarcia/go_base_project/auth"
)

func RegisterClientPasswordAuth(r *gin.Engine) {
	authGroup := r.Group("/auth")
	authGroup.POST("/login", func(c *gin.Context) {
		var body struct {
			Email    string ` + "`json:\"email\"`" + `
			Password string ` + "`json:\"password\"`" + `
		}

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, gin.H{"error": "Invalid payload"})
			return
		}

		// TODO: validate user credentials against the database
		if body.Email != "demo@example.com" || body.Password != "password123" {
			c.JSON(401, gin.H{"error": "Invalid credentials"})
			return
		}

		token, err := auth.IssueClientToken(body.Email)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(200, gin.H{"access_token": token, "token_type": "Bearer"})
	})
}`
}

func createAuthMigration(flow string) error {
	var migrationName, content string
	timestamp := time.Now().Format("20060102_150405")

	switch flow {
	case "client_password":
		migrationName = "create_users_for_auth"
		content = fmt.Sprintf(`package migrations

import (
	"time"
	"gorm.io/gorm"
)

type authUserModel struct {
	ID        uint   `+"`gorm:\"primaryKey\"`"+`
	Name      string
	Email     string `+"`gorm:\"uniqueIndex\"`"+`
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (authUserModel) TableName() string {
	return "users"
}

type CreateUsersForAuth struct{}

func (m CreateUsersForAuth) Name() string {
	return "%s_create_users_for_auth"
}

func (m CreateUsersForAuth) Up(db *gorm.DB) error {
	return db.AutoMigrate(&authUserModel{})
}

func (m CreateUsersForAuth) Down(db *gorm.DB) error {
	return db.Migrator().DropTable("users")
}

func init() {
	Register(CreateUsersForAuth{})
}
`, timestamp)

	case "client_credentials":
		migrationName = "create_oauth_clients"
		content = fmt.Sprintf(`package migrations

import (
	"time"
	"gorm.io/gorm"
)

type oauthClientModel struct {
	ID           uint   `+"`gorm:\"primaryKey\"`"+`
	ClientID     string `+"`gorm:\"uniqueIndex\"`"+`
	ClientSecret string
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (oauthClientModel) TableName() string {
	return "oauth_clients"
}

type CreateOAuthClients struct{}

func (m CreateOAuthClients) Name() string {
	return "%s_create_oauth_clients"
}

func (m CreateOAuthClients) Up(db *gorm.DB) error {
	return db.AutoMigrate(&oauthClientModel{})
}

func (m CreateOAuthClients) Down(db *gorm.DB) error {
	return db.Migrator().DropTable("oauth_clients")
}

func init() {
	Register(CreateOAuthClients{})
}
`, timestamp)

	default:
		return nil
	}

	filename := fmt.Sprintf("migrations/%s_%s.go", timestamp, migrationName)
	return os.WriteFile(filename, []byte(content), 0644)
}

func generateKeys() error {
	privCmd := `openssl genpkey -algorithm RSA -out keys/private.pem -pkeyopt rsa_keygen_bits:2048`
	pubCmd := `openssl rsa -in keys/private.pem -pubout -out keys/public.pem`

	if err := runShell(privCmd); err != nil {
		return err
	}
	if err := runShell(pubCmd); err != nil {
		return err
	}
	return nil
}

func runShell(command string) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("💻 Executing: %s\n", command)
	return cmd.Run()
}
