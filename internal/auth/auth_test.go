package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenGenerationAndValidation(t *testing.T) {
	userID := 123
	role := "admin"
	ip := "192.168.1.1"
	ua := "Mozilla/5.0"

	token, err := GenerateToken(userID, role, ip, ua)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, role, claims.Role)
	
	fingerprint := GenerateClientFingerprint(ip, ua)
	assert.Equal(t, fingerprint, claims.ClientFingerprint)
}

func TestPasswordHashing(t *testing.T) {
	password := "super-secure-password"
	
	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEqual(t, password, hash)

	assert.True(t, CheckPasswordHash(password, hash))
	assert.False(t, CheckPasswordHash("wrong-password", hash))
}

func TestGenerateClientFingerprint_IPV6(t *testing.T) {
	ip := "[::1]:1234"
	ua := "Go-Client"
	
	fingerprint1 := GenerateClientFingerprint(ip, ua)
	fingerprint2 := GenerateClientFingerprint("[::1]:5678", ua) // Different port, same IP
	
	assert.Equal(t, fingerprint1, fingerprint2)
}
