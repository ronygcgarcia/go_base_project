package auth

import (
	"crypto/rsa"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var privateKey *rsa.PrivateKey

func loadPrivateKey() (*rsa.PrivateKey, error) {
	if privateKey != nil {
		return privateKey, nil
	}

	keyData, err := os.ReadFile("keys/private.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	privateKey = privKey
	return privateKey, nil
}

func IssueClientToken(clientID string) (string, error) {
	key, err := loadPrivateKey()
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"sub":   clientID,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour * 1).Unix(),
		"scope": "client",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(key)
}

func IssueUserToken(userID uint) (string, error) {
	key, err := loadPrivateKey()
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "go_base_project"
	}
	if err != nil {
		return "", err
	}

	expiration := getTokenExpirationMinutes(60)

	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("user:%d", userID),
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(expiration).Unix(),
		"scope": "user",
		"aud":   appName,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(key)
}

func getTokenExpirationMinutes(defaultValue int) time.Duration {
	val := os.Getenv("TOKEN_EXPIRATION_MINUTES")
	if val == "" {
		return time.Duration(defaultValue) * time.Minute
	}
	if min, err := strconv.Atoi(val); err == nil {
		return time.Duration(min) * time.Minute
	}
	return time.Duration(defaultValue) * time.Minute
}
