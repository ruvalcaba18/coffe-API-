package middleware

import (
	"log"
	"net/http"
	"os"
)

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy",
			"default-src 'none'; script-src 'none'; object-src 'none'; frame-ancestors 'none'")
		if os.Getenv("ENV") == "production" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		w.Header().Del("Server")
		w.Header().Del("X-Powered-By")

		next.ServeHTTP(w, r)
	})
}

func ValidateJWTSecret() {
	env := os.Getenv("ENV")
	secret := os.Getenv("JWT_SECRET")
	if env == "production" && (secret == "" || secret == "my-secret-key-12345") {
		log.Fatal("[SECURITY] JWT_SECRET must be set to a strong random value in production. Refusing to start.")
	}
	if secret == "" {
		log.Println("[SECURITY WARNING] JWT_SECRET is not set — using insecure default. Set it before deploying to production.")
	}
}
