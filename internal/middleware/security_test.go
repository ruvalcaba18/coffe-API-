package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders_SetsAllHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		responseWriter.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	assert.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", recorder.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", recorder.Header().Get("X-XSS-Protection"))
	assert.Contains(t, recorder.Header().Get("Content-Security-Policy"), "default-src 'none'")
	assert.Equal(t, "strict-origin-when-cross-origin", recorder.Header().Get("Referrer-Policy"))
	assert.Contains(t, recorder.Header().Get("Permissions-Policy"), "geolocation=()")
}

func TestSecurityHeaders_HSTSOnlyInProduction(t *testing.T) {
	// Without production env — no HSTS
	os.Unsetenv("ENV")
	handler := SecurityHeaders(http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		responseWriter.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	assert.Empty(t, recorder.Header().Get("Strict-Transport-Security"))

	// With production env — HSTS present
	os.Setenv("ENV", "production")
	defer os.Unsetenv("ENV")

	request = httptest.NewRequest(http.MethodGet, "/", nil)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	assert.Contains(t, recorder.Header().Get("Strict-Transport-Security"), "max-age=31536000")
}
