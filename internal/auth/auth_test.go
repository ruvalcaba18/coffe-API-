package auth

import (
	"coffeebase-api/internal/models/user"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenGenerationAndValidation(t *testing.T) {
	userID := 123
	role := string(user.RoleAdmin)
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	token, error := GenerateToken(userID, role, ipAddress, userAgent)
	assert.NoError(t, error)
	assert.NotEmpty(t, token)

	claims, error := ValidateToken(token)
	assert.NoError(t, error)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, role, claims.Role)
	
	fingerprint := GenerateClientFingerprint(ipAddress, userAgent)
	assert.Equal(t, fingerprint, claims.ClientFingerprint)
}

func TestPasswordHashing(t *testing.T) {
	password := "super-secure-password"
	
	hash, error := HashPassword(password)
	assert.NoError(t, error)
	assert.NotEqual(t, password, hash)

	assert.True(t, CheckPasswordHash(password, hash))
	assert.False(t, CheckPasswordHash("wrong-password", hash))
}

func TestGenerateClientFingerprint_IPV6(t *testing.T) {
	ipAddress := "[::1]:1234"
	userAgent := "Go-Client"
	
	fingerprint1 := GenerateClientFingerprint(ipAddress, userAgent)
	fingerprint2 := GenerateClientFingerprint("[::1]:5678", userAgent) 
	assert.Equal(t, fingerprint1, fingerprint2)
}
