package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var publicKey []byte

func loadPublicKey() ([]byte, error) {
	if publicKey != nil {
		return publicKey, nil
	}
	key, err := os.ReadFile("keys/public.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}
	publicKey = key
	return publicKey, nil
}

func AuthRequired(expectedScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		key, err := loadPublicKey()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to load public key"})
			return
		}

		parsedKey, err := jwt.ParseRSAPublicKeyFromPEM(key)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid public key"})
			return
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return parsedKey, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["scope"] != expectedScope {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient scope"})
			return
		}

		c.Next()
	}
}
