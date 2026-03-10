package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func getSecretKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte("my-secret-key-12345")
	}
	return []byte(secret)
}

type Claims struct {
	UserID            int    `json:"user_id"`
	Role              string `json:"role"`
	ClientFingerprint string `json:"client_fingerprint"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int, role string, clientIP string, userAgent string) (string, error) {
	expirationDuration := 2 * time.Hour
	tokenExpirationTime := time.Now().Add(expirationDuration)
	
	uniqueClientFingerprint := GenerateClientFingerprint(clientIP, userAgent)
	
	sessionIdentifier := uuid.New().String()

	tokenClaims := &Claims{
		UserID:            userID,
		Role:              role,
		ClientFingerprint: uniqueClientFingerprint,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        sessionIdentifier,
			ExpiresAt: jwt.NewNumericDate(tokenExpirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	signedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	return signedToken.SignedString(getSecretKey())
}

func ValidateToken(tokenString string) (*Claims, error) {
	tokenClaims := &Claims{}
	parsedToken, parsingError := jwt.ParseWithClaims(tokenString, tokenClaims, func(token *jwt.Token) (interface{}, error) {
		return getSecretKey(), nil
	})

	if parsingError != nil {
		return nil, parsingError
	}

	if !parsedToken.Valid {
		return nil, errors.New("the provided token is no longer valid")
	}

	return tokenClaims, nil
}

func GenerateClientFingerprint(ipAddress string, userAgent string) string {

	normalizedIP := ipAddress
	for i := len(ipAddress) - 1; i >= 0; i-- {
		if ipAddress[i] == ':' {
			normalizedIP = ipAddress[:i]
			break
		}
	}

	combinedAttributes := fmt.Sprintf("%s|%s", normalizedIP, userAgent)
	attributeHash := sha256.Sum256([]byte(combinedAttributes))
	return hex.EncodeToString(attributeHash[:])
}

func HashPassword(rawPassword string) (string, error) {
	passwordBytes, hashingError := bcrypt.GenerateFromPassword([]byte(rawPassword), 14)
	return string(passwordBytes), hashingError
}

func CheckPasswordHash(rawPassword string, storedHash string) bool {
	comparisonError := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(rawPassword))
	return comparisonError == nil
}
