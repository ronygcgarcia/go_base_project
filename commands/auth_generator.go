package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
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

	createOauthClientCommand()
	addCreateOauthClientCommandCli()
	if err := addCreateOauthClientCommandCli(); err != nil {
		return fmt.Errorf("failed to add CLI command: %w", err)
	}
	fmt.Println("✅ Auth flow setup completed successfully.")

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
	authGroup.POST("/refresh", controllers.RefreshAccessToken)
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

	// Crea el archivo si no existe
	if _, err := os.Stat(controllerFile); os.IsNotExist(err) {
		header := `package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/ronygcgarcia/go_base_project/auth"
	"github.com/ronygcgarcia/go_base_project/config"
	"github.com/ronygcgarcia/go_base_project/models"
	"time"
	"os"
	"strconv"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)
`
		if err := os.WriteFile(controllerFile, []byte(header), 0644); err != nil {
			return fmt.Errorf("failed to create controller file: %w", err)
		}
		fmt.Println("✅ Created new controller file:", controllerFile)
	}

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

	if !client.CheckSecret(body.ClientSecret) || client.Revoked {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	accessToken, err := auth.IssueClientToken(client.ClientID, getAccessTokenTTL())
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken := uuid.New().String()
	hashed, _ := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)

	rt := models.RefreshToken{
		ClientID:  client.ID,
		Token:     string(hashed),
		ExpiresAt: time.Now().Add(getRefreshTTL()),
	}
	config.DB.Create(&rt)

	c.JSON(200, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    int(getAccessTokenTTL().Seconds()),
	})
}

func RefreshAccessToken(c *gin.Context) {
	var body struct {
		RefreshToken string ` + "`json:\"refresh_token\"`" + `
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid payload"})
		return
	}

	var token models.RefreshToken
	if err := config.DB.Where("revoked = ?", false).Find(&token).Error; err != nil || token.ID == 0 {
		c.JSON(401, gin.H{"error": "Refresh token not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(token.Token), []byte(body.RefreshToken)); err != nil ||
		token.ExpiresAt.Before(time.Now()) {
		c.JSON(401, gin.H{"error": "Refresh token is invalid or expired"})
		return
	}

	token.Revoked = true
	config.DB.Save(&token)

	var client models.OAuthClient
	config.DB.First(&client, token.ClientID)

	newAccessToken, err := auth.IssueClientToken(client.ClientID, getAccessTokenTTL())
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate new access token"})
		return
	}

	newRefresh := uuid.New().String()
	hashedNew, _ := bcrypt.GenerateFromPassword([]byte(newRefresh), bcrypt.DefaultCost)

	newRT := models.RefreshToken{
		ClientID:  client.ID,
		Token:     string(hashedNew),
		ExpiresAt: time.Now().Add(getRefreshTTL()),
	}
	config.DB.Create(&newRT)

	c.JSON(200, gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefresh,
		"token_type":    "Bearer",
		"expires_in":    int(getAccessTokenTTL().Seconds()),
	})
}`
	default:
		return fmt.Errorf("unknown auth flow: %s", flow)
	}

	// Añadir método si no está
	source, err := os.ReadFile(controllerFile)
	if err != nil {
		return err
	}
	if !strings.Contains(string(source), methodNeeded) {
		f, err := os.OpenFile(controllerFile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.WriteString(methodCode); err != nil {
			return err
		}
		fmt.Println("✅ Added method to controller:", methodNeeded)
	} else {
		fmt.Println("ℹ️ Controller already has:", methodNeeded)
	}

	// Añadir helpers TTL si no existen
	if !strings.Contains(string(source), "func getAccessTokenTTL") {
		helpers := `

func getRefreshTTL() time.Duration {
	val := os.Getenv("REFRESH_TOKEN_EXPIRATION_MINUTES")
	if min, err := strconv.Atoi(val); err == nil && min > 0 {
		return time.Duration(min) * time.Minute
	}
	return 7 * 24 * time.Hour
}

func getAccessTokenTTL() time.Duration {
	val := os.Getenv("TOKEN_EXPIRATION_MINUTES")
	if min, err := strconv.Atoi(val); err == nil && min > 0 {
		return time.Duration(min) * time.Minute
	}
	return time.Hour
}`
		f, err := os.OpenFile(controllerFile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.WriteString(helpers); err != nil {
			return err
		}
		fmt.Println("✅ Added token TTL helpers to controller")
	} else {
		fmt.Println("ℹ️ TTL helpers already exist in controller")
	}

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
	ExpireAt  *time.Time
	Revoked   bool       ` + "`gorm:\"default:false\"`" + `
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
				!strings.Contains(string(source), "UpdatedAt") ||
				!strings.Contains(string(source), "RefreshTokens") {

				appendContent := `

// Ensure full OAuthClient model compliance for client auth
import "time"

type OAuthClient struct {
	ID            uint      ` + "`gorm:\"primaryKey\"`" + `
	ClientID      string    ` + "`gorm:\"uniqueIndex\"`" + `
	ClientSecret  string
	Name          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ExpireAt      *time.Time
	Revoked       bool      ` + "`gorm:\"default:false\"`" + `
	RefreshTokens []RefreshToken ` + "`gorm:\"foreignKey:ClientID\"`" + `
}

func (OAuthClient) TableName() string {
	return "oauth_clients"
}

func (c *OAuthClient) CheckSecret(secret string) bool {
	return bcrypt.CompareHashAndPassword([]byte(c.ClientSecret), []byte(secret)) == nil
}

type RefreshToken struct {
	ID        uint      ` + "`gorm:\"primaryKey\"`" + `
	ClientID  uint      ` + "`gorm:\"index\"`" + `
	Token     string    ` + "`gorm:\"uniqueIndex\"`" + `
	ExpiresAt time.Time
	Revoked   bool      ` + "`gorm:\"default:false\"`" + `
	CreatedAt time.Time
	UpdatedAt time.Time
}`
				f, err := os.OpenFile(modelFile, os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					return err
				}
				defer f.Close()
				if _, err = f.WriteString(appendContent); err != nil {
					return err
				}
				fmt.Println("✅ Added OAuthClient + RefreshToken model to:", modelFile)
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
	ID            uint      ` + "`gorm:\"primaryKey\"`" + `
	ClientID      string    ` + "`gorm:\"uniqueIndex\"`" + `
	ClientSecret  string
	Name          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ExpireAt      *time.Time
	Revoked       bool      ` + "`gorm:\"default:false\"`" + `
	RefreshTokens []RefreshToken ` + "`gorm:\"foreignKey:ClientID\"`" + `
}

func (OAuthClient) TableName() string {
	return "oauth_clients"
}

func (c *OAuthClient) CheckSecret(secret string) bool {
	return bcrypt.CompareHashAndPassword([]byte(c.ClientSecret), []byte(secret)) == nil
}

type RefreshToken struct {
	ID        uint      ` + "`gorm:\"primaryKey\"`" + `
	ClientID  uint      ` + "`gorm:\"index\"`" + `
	Token     string    ` + "`gorm:\"uniqueIndex\"`" + `
	ExpiresAt time.Time
	Revoked   bool      ` + "`gorm:\"default:false\"`" + `
	CreatedAt time.Time
	UpdatedAt time.Time
}`
	}

	return os.WriteFile(modelFile, []byte(modelContent), 0644)
}

func createAuthMigration(flow string) error {
	var migrationName, content string

	switch flow {
	case "client_password":
		migrationName = "create_users_for_auth"
		if _, err := os.Stat("migrations/000001_" + migrationName + ".go"); err == nil {
			fmt.Println("ℹ️ Migration already exists:", migrationName)
			return nil
		}
		content = `package migrations

import (
	"time"
	"gorm.io/gorm"
)

type authUserModel struct {
	ID        uint      ` + "`gorm:\"primaryKey\"`" + `
	Name      string
	Email     string     ` + "`gorm:\"uniqueIndex\"`" + `
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpireAt  *time.Time
	Revoked   bool       ` + "`gorm:\"default:false\"`" + `
}

func (authUserModel) TableName() string {
	return "users"
}

type CreateUsersForAuth struct{}

func (m CreateUsersForAuth) Name() string {
	return "000001_create_users_for_auth"
}

func (m CreateUsersForAuth) Up(db *gorm.DB) error {
	return db.AutoMigrate(&authUserModel{})
}

func (m CreateUsersForAuth) Down(db *gorm.DB) error {
	return db.Migrator().DropTable("users")
}

func init() {
	Register(CreateUsersForAuth{})
}`
		filename := "migrations/000001_create_users_for_auth.go"
		return os.WriteFile(filename, []byte(content), 0644)

	case "client_credentials":
		// oauth_clients migration
		migrationName = "create_oauth_clients"
		if _, err := os.Stat("migrations/000002_" + migrationName + ".go"); err == nil {
			fmt.Println("ℹ️ Migration already exists:", migrationName)
		} else {
			content = `package migrations

import (
	"time"
	"gorm.io/gorm"
)

type oauthClientModel struct {
	ID           uint      ` + "`gorm:\"primaryKey\"`" + `
	ClientID     string    ` + "`gorm:\"uniqueIndex\"`" + `
	ClientSecret string
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ExpireAt     *time.Time
	Revoked      bool      ` + "`gorm:\"default:false\"`" + `
}

func (oauthClientModel) TableName() string {
	return "oauth_clients"
}

type CreateOAuthClients struct{}

func (m CreateOAuthClients) Name() string {
	return "000002_create_oauth_clients"
}

func (m CreateOAuthClients) Up(db *gorm.DB) error {
	return db.AutoMigrate(&oauthClientModel{})
}

func (m CreateOAuthClients) Down(db *gorm.DB) error {
	return db.Migrator().DropTable("oauth_clients")
}

func init() {
	Register(CreateOAuthClients{})
}`
			filename := "migrations/000002_create_oauth_clients.go"
			if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
				return err
			}
			fmt.Println("✅ Created migration:", filename)
		}

		// refresh_tokens migration
		migrationName = "create_refresh_tokens"
		if _, err := os.Stat("migrations/000003_" + migrationName + ".go"); err == nil {
			fmt.Println("ℹ️ Migration already exists:", migrationName)
			return nil
		}
		content = `package migrations

import (
	"time"
	"gorm.io/gorm"
)

type refreshTokenModel struct {
	ID             uint      ` + "`gorm:\"primaryKey\"`" + `
	ClientID       uint      ` + "`gorm:\"index\"`" + `
	Token          string    ` + "`gorm:\"uniqueIndex\"`" + `
	ExpiresAt      time.Time
	Revoked        bool      ` + "`gorm:\"default:false\"`" + `
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (refreshTokenModel) TableName() string {
	return "refresh_tokens"
}

type CreateRefreshTokens struct{}

func (m CreateRefreshTokens) Name() string {
	return "000003_create_refresh_tokens"
}

func (m CreateRefreshTokens) Up(db *gorm.DB) error {
	return db.AutoMigrate(&refreshTokenModel{})
}

func (m CreateRefreshTokens) Down(db *gorm.DB) error {
	return db.Migrator().DropTable("refresh_tokens")
}

func init() {
	Register(CreateRefreshTokens{})
}`
		filename := "migrations/000003_create_refresh_tokens.go"
		return os.WriteFile(filename, []byte(content), 0644)
	}

	return nil
}

func createOauthClientCommand() {
	// Create commands/create_oauth_client.go if missing
	oauthCmdPath := "commands/create_oauth_client.go"
	if _, err := os.Stat(oauthCmdPath); os.IsNotExist(err) {
		oauthContent := `package commands

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ronygcgarcia/go_base_project/config"
	"github.com/ronygcgarcia/go_base_project/models"
	"golang.org/x/crypto/bcrypt"
)

func CreateOAuthClientCLI(name string, minutes int) error {
	clientID, err := generateSecureRandomHex(16)
	if err != nil {
		return fmt.Errorf("failed to generate client ID: %w", err)
	}

	clientSecretRaw, err := generateSecureRandomHex(32)
	if err != nil {
		return fmt.Errorf("failed to generate client secret: %w", err)
	}

	hashedSecret, _ := bcrypt.GenerateFromPassword([]byte(clientSecretRaw), bcrypt.DefaultCost)

	client := models.OAuthClient{
		ClientID:     clientID,
		ClientSecret: string(hashedSecret),
		Name:         name,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Revoked:      false,
	}

	if minutes > 0 {
		exp := time.Now().Add(time.Duration(minutes) * time.Minute)
		client.ExpireAt = &exp
	}

	if err := config.DB.Create(&client).Error; err != nil {
		return fmt.Errorf("failed to save oauth client: %w", err)
	}

	fmt.Println("✅ OAuth client created successfully")
	fmt.Println("Client ID:     ", clientID)
	fmt.Println("Client Secret: ", clientSecretRaw)
	return nil
}

func generateSecureRandomHex(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
`
		if err := os.WriteFile(oauthCmdPath, []byte(oauthContent), 0644); err != nil {
			fmt.Printf("❌ Error: failed to create CLI file: %v\n", err)
			return
		}
		fmt.Println("✅ Created command:", oauthCmdPath)
	} else {
		fmt.Println("ℹ️ CLI file already exists:", oauthCmdPath)
	}

}

func addCreateOauthClientCommandCli() error {
	cliPath := "commands/cli.go"
	cliBytes, err := os.ReadFile(cliPath)
	if err != nil {
		return fmt.Errorf("could not read cli.go: %w", err)
	}
	cliContent := string(cliBytes)

	// Agregar import "strconv" si no está
	if !strings.Contains(cliContent, "\"strconv\"") {
		cliContent = strings.Replace(cliContent, "import (", "import (\n\t\"strconv\"", 1)
	}

	caseString := `case "create:oauth-client-credentials":`

	if !strings.Contains(cliContent, caseString) {
		inject := `
	case "create:oauth-client-credentials":
		config.ConnectDatabase()
		var name string
		var minutes int
		for _, arg := range os.Args[2:] {
			if strings.HasPrefix(arg, "--name=") {
				name = strings.TrimPrefix(arg, "--name=")
			} else if strings.HasPrefix(arg, "--expires=") {
				expiresStr := strings.TrimPrefix(arg, "--expires=")
				minutes, _ = strconv.Atoi(expiresStr)
			}
		}
		if name == "" {
			fmt.Println("❌ --name parameter is required")
			os.Exit(1)
		}
		if err := CreateOAuthClientCLI(name, minutes); err != nil {
			fmt.Println("❌ Error creating oauth client:", err)
			os.Exit(1)
		}
		return true`

		newCli := strings.Replace(cliContent, "switch cmd {", "switch cmd {"+inject, 1)
		if err := os.WriteFile(cliPath, []byte(newCli), 0644); err != nil {
			return fmt.Errorf("failed to patch cli.go: %w", err)
		}
		fmt.Println("✅ Patched cli.go to include oauth-client-credentials command")
	} else {
		fmt.Println("ℹ️ CLI command already present in cli.go")
	}

	return nil
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
