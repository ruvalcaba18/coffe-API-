package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID            int    `json:"user_id"`
	Role              string `json:"role"`
	ClientFingerprint string `json:"client_fingerprint"`
	jwt.RegisteredClaims
}

// --- Public ---

func GenerateToken(userID int, role string, clientIP string, userAgent string) (string, error) {
	expiration := 2 * time.Hour
	expiresAt := time.Now().Add(expiration)
	fingerprint := GenerateClientFingerprint(clientIP, userAgent)
	sessionID := uuid.New().String()

	claims := &Claims{
		UserID:            userID,
		Role:              role,
		ClientFingerprint: fingerprint,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        sessionID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSecretKey())
}

func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, error := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return getSecretKey(), nil
	})

	if error != nil {
		return nil, error
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func GenerateClientFingerprint(ip string, userAgent string) string {
	normalizedIP := ip
	if host, _, error := net.SplitHostPort(ip); error == nil {
		normalizedIP = host
	}
	normalizedIP = strings.TrimPrefix(normalizedIP, "[")
	normalizedIP = strings.TrimSuffix(normalizedIP, "]")

	hash := sha256.Sum256([]byte(fmt.Sprintf("%s|%s", normalizedIP, userAgent)))
	return hex.EncodeToString(hash[:])
}

func HashPassword(password string) (string, error) {
	bytes, error := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), error
}

func CheckPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// --- Private ---

func getSecretKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte("my-secret-key-12345")
	}
	return []byte(secret)
}
