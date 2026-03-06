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

var cryptographySecretKey = []byte(os.Getenv("JWT_SECRET"))

func init() {
	if len(cryptographySecretKey) == 0 {
		cryptographySecretKey = []byte("my-secret-key-12345")
	}
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
	return signedToken.SignedString(cryptographySecretKey)
}

func ValidateToken(tokenString string) (*Claims, error) {
	tokenClaims := &Claims{}
	parsedToken, parsingError := jwt.ParseWithClaims(tokenString, tokenClaims, func(token *jwt.Token) (interface{}, error) {
		return cryptographySecretKey, nil
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
	combinedAttributes := fmt.Sprintf("%s|%s", ipAddress, userAgent)
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
