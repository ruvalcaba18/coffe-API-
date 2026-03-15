package middleware

import (
	"coffeebase-api/internal/auth"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		userID := httpRequest.Context().Value(UserIDKey).(int)
		role := httpRequest.Context().Value(UserRoleKey).(string)
		assert.Equal(t, 1, userID)
		assert.Equal(t, "customer", role)
		responseWriter.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(nextHandler)

	t.Run("Valid Token in Header", func(t *testing.T) {
		ipAddress := "127.0.0.1"
		userAgent := "Go-Test-Agent"
		token, _ := auth.GenerateToken(1, "customer", ipAddress, userAgent)

		httpRequest := httptest.NewRequest("GET", "/", nil)
		httpRequest.Header.Set("Authorization", "Bearer "+token)
		httpRequest.Header.Set("User-Agent", userAgent)
		httpRequest.RemoteAddr = ipAddress + ":1234" // Simulate port

		responseRecorder := httptest.NewRecorder()
		middleware.ServeHTTP(responseRecorder, httpRequest)

		assert.Equal(t, http.StatusOK, responseRecorder.Code)
	})

	t.Run("Missing Token", func(t *testing.T) {
		httpRequest := httptest.NewRequest("GET", "/", nil)
		responseRecorder := httptest.NewRecorder()
		middleware.ServeHTTP(responseRecorder, httpRequest)

		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		httpRequest := httptest.NewRequest("GET", "/", nil)
		httpRequest.Header.Set("Authorization", "Bearer invalid-token")
		responseRecorder := httptest.NewRecorder()
		middleware.ServeHTTP(responseRecorder, httpRequest)

		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)
	})

	t.Run("Fingerprint Mismatch", func(t *testing.T) {
		token, _ := auth.GenerateToken(1, "customer", "1.1.1.1", "Browser 1")

		httpRequest := httptest.NewRequest("GET", "/", nil)
		httpRequest.Header.Set("Authorization", "Bearer "+token)
		httpRequest.Header.Set("User-Agent", "Browser 2") // Different UA
		httpRequest.RemoteAddr = "2.2.2.2:80"             // Different IP

		responseRecorder := httptest.NewRecorder()
		middleware.ServeHTTP(responseRecorder, httpRequest)

		assert.Equal(t, http.StatusForbidden, responseRecorder.Code)
	})
}
