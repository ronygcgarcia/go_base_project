package commands

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
	// Generar client_id aleatorio (hex de 16 bytes)
	clientID, err := generateSecureRandomHex(16)
	if err != nil {
		return fmt.Errorf("failed to generate client ID: %w", err)
	}

	// Generar client_secret sin hash para mostrar y luego hashearlo para guardar
	clientSecretRaw, err := generateSecureRandomHex(32)
	if err != nil {
		return fmt.Errorf("failed to generate client secret: %w", err)
	}

	hashedSecret, _ := bcrypt.GenerateFromPassword([]byte(clientSecretRaw), bcrypt.DefaultCost)

	// Armar struct del cliente OAuth
	oauthClient := models.OAuthClient{
		ClientID:     clientID,
		ClientSecret: string(hashedSecret),
		Name:         name,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Revoked:      false,
	}

	// Agregar expiración si se proporcionó
	if minutes > 0 {
		exp := time.Now().Add(time.Duration(minutes) * time.Minute)
		oauthClient.ExpireAt = &exp
	}

	// Guardar en base de datos
	if err := config.DB.Create(&oauthClient).Error; err != nil {
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
