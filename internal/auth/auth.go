package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
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
	token, error := jwt.ParseWithClaims(tokenString, claims, jwtKeyLookup)

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

// GenerateRefreshToken genera un token opaco criptográficamente seguro.
// Devuelve (rawToken, tokenHash, error).
// rawToken: string para enviar al cliente (nunca se persiste en DB).
// tokenHash: HMAC-SHA256 del rawToken, lo que SÍ se almacena en DB.
func GenerateRefreshToken() (rawToken string, tokenHash string, err error) {
	bytes := make([]byte, 32)
	if _, err = rand.Read(bytes); err != nil {
		return "", "", err
	}
	rawToken = base64.URLEncoding.EncodeToString(bytes)
	tokenHash = HashRefreshToken(rawToken)
	return rawToken, tokenHash, nil
}

// HashRefreshToken calcula el HMAC-SHA256 de un refresh token raw.
// Usa REFRESH_TOKEN_SECRET (env) como clave — separado del JWT_SECRET.
func HashRefreshToken(rawToken string) string {
	secret := os.Getenv("REFRESH_TOKEN_SECRET")
	if secret == "" {
		secret = "refresh-secret-change-me"
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(rawToken))
	return hex.EncodeToString(mac.Sum(nil))
}

// GenerateTokenPair genera el par access_token + refresh_token para una sesión.
// Devuelve (accessToken, rawRefreshToken, error).
func GenerateTokenPair(userID int, role, ip, userAgent string) (accessToken, rawRefreshToken string, err error) {
	accessToken, err = GenerateToken(userID, role, ip, userAgent)
	if err != nil {
		return "", "", err
	}
	rawRefreshToken, _, err = GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	return accessToken, rawRefreshToken, nil
}

// --- Private ---

func jwtKeyLookup(token *jwt.Token) (interface{}, error) {
	return getSecretKey(), nil
}

func getSecretKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte("my-secret-key-12345")
	}
	return []byte(secret)
}
