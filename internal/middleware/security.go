package middleware

import (
	"log/slog"
	"net/http"
	"os"
)

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		responseWriter.Header().Set("X-Content-Type-Options", "nosniff")
		responseWriter.Header().Set("X-Frame-Options", "DENY")
		responseWriter.Header().Set("X-XSS-Protection", "1; mode=block")
		responseWriter.Header().Set("Content-Security-Policy",
			"default-src 'none'; script-src 'none'; object-src 'none'; frame-ancestors 'none'")
		if os.Getenv("ENV") == "production" {
			responseWriter.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		responseWriter.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		responseWriter.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		responseWriter.Header().Del("Server")
		responseWriter.Header().Del("X-Powered-By")

		// OWASP A05: Evitar cacheo de respuestas sensibles de la API
		responseWriter.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		responseWriter.Header().Set("Pragma", "no-cache")

		next.ServeHTTP(responseWriter, httpRequest)
	})
}

func ValidateJWTSecret() {
	environment := os.Getenv("ENV")
	jwtSecret := os.Getenv("JWT_SECRET")
	if environment == "production" && (jwtSecret == "" || jwtSecret == "my-secret-key-12345") {
		slog.Error("[SECURITY] JWT_SECRET must be set to a strong random value in production. Refusing to start.")
		os.Exit(1)
	}
	if jwtSecret == "" {
		slog.Warn("[SECURITY WARNING] JWT_SECRET is not set — using insecure default. Set it before deploying to production.")
	}

	// OWASP A02: Verificar también los secrets de refresh token y QR
	if refreshSecret := os.Getenv("REFRESH_TOKEN_SECRET"); refreshSecret == "" {
		slog.Warn("[SECURITY WARNING] REFRESH_TOKEN_SECRET is not set — using insecure default. Set it before deploying to production.")
	}
	if qrSecret := os.Getenv("QR_TOKEN_SECRET"); qrSecret == "" {
		slog.Warn("[SECURITY WARNING] QR_TOKEN_SECRET is not set — using insecure default. Set it before deploying to production.")
	}
}
