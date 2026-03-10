package middleware

import (
	"coffeebase-api/internal/auth"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDKey).(int)
		role := r.Context().Value(UserRoleKey).(string)
		assert.Equal(t, 1, userID)
		assert.Equal(t, "customer", role)
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(nextHandler)

	t.Run("Valid Token in Header", func(t *testing.T) {
		ip := "127.0.0.1"
		ua := "Go-Test-Agent"
		token, _ := auth.GenerateToken(1, "customer", ip, ua)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", ua)
		req.RemoteAddr = ip + ":1234" // Simulate port

		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Missing Token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Fingerprint Mismatch", func(t *testing.T) {
		token, _ := auth.GenerateToken(1, "customer", "1.1.1.1", "Browser 1")

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", "Browser 2") // Different UA
		req.RemoteAddr = "2.2.2.2:80"             // Different IP

		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})
}
