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

	// Generate controller and model
	if err := generateAuthController(flow); err != nil {
		return fmt.Errorf("failed to create controller: %w", err)
	}
	if err := generateAuthModel(flow); err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	return nil
}

func getClientCredentialsTemplate() string {
	return `package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ronygcgarcia/go_base_project/controllers"
)

func RegisterClientCredentialsAuth(r *gin.Engine) {
	authGroup := r.Group("/auth")
	authGroup.POST("/token", controllers.AuthClientCredentials)
}`
}

func getClientPasswordTemplate() string {
	return `package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ronygcgarcia/go_base_project/controllers"
)

func RegisterClientPasswordAuth(r *gin.Engine) {
	authGroup := r.Group("/auth")
	authGroup.POST("/login", controllers.LoginUser)
}`
}

// generateAuthController creates a basic auth controller file with login logic.
func generateAuthController(flow string) error {
	controllerFile := "controllers/auth.controller.go"
	methodNeeded := ""
	methodCode := ""

	switch flow {
	case "client_password":
		methodNeeded = "func LoginUser"
		methodCode = `

func LoginUser(c *gin.Context) {
	var body struct {
		Email    string ` + "`json:\"email\"`" + `
		Password string ` + "`json:\"password\"`" + `
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", body.Email).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.CheckPassword(body.Password) {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := auth.IssueUserToken(user.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(200, gin.H{"access_token": token, "token_type": "Bearer"})
}`

	case "client_credentials":
		methodNeeded = "func AuthClientCredentials"
		methodCode = `

func AuthClientCredentials(c *gin.Context) {
	var body struct {
		ClientID     string ` + "`json:\"client_id\"`" + `
		ClientSecret string ` + "`json:\"client_secret\"`" + `
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	var client models.OAuthClient
	if err := config.DB.Where("client_id = ?", body.ClientID).First(&client).Error; err != nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	if !client.CheckSecret(body.ClientSecret) {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := auth.IssueClientToken(client.ClientID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(200, gin.H{"access_token": token, "token_type": "Bearer"})
}`
	default:
		return fmt.Errorf("unknown auth flow: %s", flow)
	}

	// If file doesn't exist, create it
	if _, err := os.Stat(controllerFile); os.IsNotExist(err) {
		header := `package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/ronygcgarcia/go_base_project/auth"
	"github.com/ronygcgarcia/go_base_project/config"
	"github.com/ronygcgarcia/go_base_project/models"
)` + methodCode
		return os.WriteFile(controllerFile, []byte(header), 0644)
	}

	// If file exists, append method if not present
	source, err := os.ReadFile(controllerFile)
	if err != nil {
		return err
	}
	if strings.Contains(string(source), methodNeeded) {
		fmt.Println("ℹ️ Controller already has:", methodNeeded)
		return nil
	}

	f, err := os.OpenFile(controllerFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(methodCode); err != nil {
		return err
	}

	fmt.Println("✅ Added method to controller:", methodNeeded)
	return nil
}

// generateAuthModel creates the user or oauth_client model if needed.
func generateAuthModel(flow string) error {
	var modelFile, modelContent string

	switch flow {
	case "client_password":
		modelFile = "models/user.go"
		if _, err := os.Stat(modelFile); err == nil {
			source, err := os.ReadFile(modelFile)
			if err != nil {
				return err
			}
			if !strings.Contains(string(source), "CheckPassword") ||
				!strings.Contains(string(source), "Email") ||
				!strings.Contains(string(source), "Password") ||
				!strings.Contains(string(source), "Name") ||
				!strings.Contains(string(source), "CreatedAt") ||
				!strings.Contains(string(source), "UpdatedAt") {

				appendContent := `

// Ensure full User model compliance for auth
import "time"

type User struct {
	ID        uint      ` + "`gorm:\"primaryKey\"`" + `
	Name      string
	Email     string    ` + "`gorm:\"uniqueIndex\"`" + `
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (User) TableName() string {
	return "users"
}

func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}`
				f, err := os.OpenFile(modelFile, os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					return err
				}
				defer f.Close()
				if _, err = f.WriteString(appendContent); err != nil {
					return err
				}
				fmt.Println("✅ Added CheckPassword method and fields to:", modelFile)
			} else {
				fmt.Println("ℹ️ Model already has required fields:", modelFile)
			}
			return nil
		}
		modelContent = `package models

import (
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	ID        uint      ` + "`gorm:\"primaryKey\"`" + `
	Name      string
	Email     string    ` + "`gorm:\"uniqueIndex\"`" + `
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (User) TableName() string {
	return "users"
}

func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}`
	case "client_credentials":
		modelFile = "models/oauth_client.go"
		if _, err := os.Stat(modelFile); err == nil {
			source, err := os.ReadFile(modelFile)
			if err != nil {
				return err
			}
			if !strings.Contains(string(source), "CheckSecret") ||
				!strings.Contains(string(source), "ClientID") ||
				!strings.Contains(string(source), "ClientSecret") ||
				!strings.Contains(string(source), "Name") ||
				!strings.Contains(string(source), "CreatedAt") ||
				!strings.Contains(string(source), "UpdatedAt") {

				appendContent := `

// Ensure full OAuthClient model compliance for client auth
import "time"

type OAuthClient struct {
	ID           uint      ` + "`gorm:\"primaryKey\"`" + `
	ClientID     string    ` + "`gorm:\"uniqueIndex\"`" + `
	ClientSecret string
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (OAuthClient) TableName() string {
	return "oauth_clients"
}

func (c *OAuthClient) CheckSecret(secret string) bool {
	return bcrypt.CompareHashAndPassword([]byte(c.ClientSecret), []byte(secret)) == nil
}`
				f, err := os.OpenFile(modelFile, os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					return err
				}
				defer f.Close()
				if _, err = f.WriteString(appendContent); err != nil {
					return err
				}
				fmt.Println("✅ Added CheckSecret method and fields to:", modelFile)
			} else {
				fmt.Println("ℹ️ Model already has required fields:", modelFile)
			}
			return nil
		}
		modelContent = `package models

import (
	"golang.org/x/crypto/bcrypt"
	"time"
)

type OAuthClient struct {
	ID           uint      ` + "`gorm:\"primaryKey\"`" + `
	ClientID     string    ` + "`gorm:\"uniqueIndex\"`" + `
	ClientSecret string
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (OAuthClient) TableName() string {
	return "oauth_clients"
}

func (c *OAuthClient) CheckSecret(secret string) bool {
	return bcrypt.CompareHashAndPassword([]byte(c.ClientSecret), []byte(secret)) == nil
}`
	}
	return os.WriteFile(modelFile, []byte(modelContent), 0644)
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
	ExpireAt     *time.Time
	Revoked      bool
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
