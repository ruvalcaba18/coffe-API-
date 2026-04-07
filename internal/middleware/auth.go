package middleware

import (
	"coffeebase-api/api/response"
	"coffeebase-api/internal/auth"
	"context"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const UserRoleKey contextKey = "user_role"

// --- Public ---

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		token := retrieveToken(httpRequest)
		if token == "" {
			response.Unauthorized(responseWriter, "Authentication required")
			return
		}

		claims, error := auth.ValidateToken(token)
		if error != nil {
			response.Unauthorized(responseWriter, "Invalid or expired token")
			return
		}

		if !validateFingerprint(httpRequest, claims.ClientFingerprint) {
			response.Forbidden(responseWriter, "Security Violation: Unauthorized device signature")
			return
		}

		requestContext := context.WithValue(httpRequest.Context(), UserIDKey, claims.UserID)
		requestContext = context.WithValue(requestContext, UserRoleKey, claims.Role)
		
		next.ServeHTTP(responseWriter, httpRequest.WithContext(requestContext))
	})
}

// --- Private ---

func retrieveToken(httpRequest *http.Request) string {
	authHeader := httpRequest.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}
	return httpRequest.URL.Query().Get("token")
}

func validateFingerprint(httpRequest *http.Request, tokenFingerprint string) bool {
	ip := httpRequest.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = httpRequest.RemoteAddr
	}

	ip = strings.Split(strings.Split(ip, ",")[0], ":")[0]

	userAgent := httpRequest.Header.Get("User-Agent")
	currentFingerprint := auth.GenerateClientFingerprint(ip, userAgent)

	if tokenFingerprint == currentFingerprint {
		return true
	}

	return isLocalAddress(ip)
}

func isLocalAddress(ip string) bool {
	return strings.HasPrefix(ip, "127.0.0.1") || 
		   strings.HasPrefix(ip, "::1") || 
		   strings.HasPrefix(ip, "[::1]") ||
		   ip == "localhost"
}
