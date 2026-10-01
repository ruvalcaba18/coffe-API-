package attendance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	attendancemodel "coffeebase-api/internal/models/attendance"

	"github.com/google/uuid"
	qrcode "github.com/skip2/go-qrcode"
)

// QRPayload es la estructura codificada dentro del QR.
type QRPayload struct {
	Token     string    `json:"token"`
	UserID    int       `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

const qrTokenDuration = 24 * time.Hour // Vigencia: 1 día

// --- Public ---

// GenerateQRForUser genera un token HMAC firmado para el empleado indicado,
// construye el payload JSON, lo convierte a código QR PNG en base64,
// y devuelve el PNG y el modelo QRToken para almacenar en DB.
func GenerateQRForUser(userID int) ([]byte, *attendancemodel.QRToken, error) {
	rawToken, tokenHash, err := generateRawToken(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("error generando token QR: %w", err)
	}

	expiresAt := time.Now().Add(qrTokenDuration)

	payload := QRPayload{
		Token:     rawToken,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("error serializando payload QR: %w", err)
	}

	pngBytes, err := qrcode.Encode(string(payloadJSON), qrcode.Medium, 256)
	if err != nil {
		return nil, nil, fmt.Errorf("error generando imagen QR: %w", err)
	}

	qrToken := &attendancemodel.QRToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	return pngBytes, qrToken, nil
}

// ValidateQRToken desempaqueta el payload JSON escaneado del QR y valida vigencia.
func ValidateQRToken(payloadJSON string) (*QRPayload, error) {
	var payload QRPayload
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		return nil, errors.New("QR inválido: formato incorrecto")
	}

	if payload.Token == "" || payload.UserID <= 0 {
		return nil, errors.New("QR inválido: payload incompleto")
	}

	if time.Now().After(payload.ExpiresAt) {
		return nil, errors.New("QR expirado")
	}

	return &payload, nil
}

// HashQRToken calcula el HMAC-SHA256 del token crudo usando QR_TOKEN_SECRET.
func HashQRToken(rawToken string) string {
	secret := getQRSecret()
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(rawToken))
	return hex.EncodeToString(mac.Sum(nil))
}

// --- Private ---

func generateRawToken(userID int) (string, string, error) {
	nonce := uuid.New().String()
	rawToken := fmt.Sprintf("%d|%d|%s", userID, time.Now().UnixNano(), nonce)
	tokenHash := HashQRToken(rawToken)
	return rawToken, tokenHash, nil
}

func getQRSecret() string {
	secret := os.Getenv("QR_TOKEN_SECRET")
	if secret == "" {
		return "qr-secret-change-me"
	}
	return secret
}
