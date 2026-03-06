package middleware

import (
	"coffeebase-api/internal/auth"
	"context"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const UserRoleKey contextKey = "user_role"

/**
 * AuthMiddleware enforces strict authentication and session integrity.
 * It verifies not only the token's signature but also its bind to the specific client.
 */
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		var authenticationToken string

		// 1. Attempt retrieval from Authorization header (Standard REST)
		authorizationHeader := httpRequest.Header.Get("Authorization")
		if authorizationHeader != "" {
			headerParts := strings.Split(authorizationHeader, " ")
			if len(headerParts) == 2 && headerParts[0] == "Bearer" {
				authenticationToken = headerParts[1]
			}
		}

		// 2. Fallback to query parameter (Standard for WebSocket initial handshakes)
		if authenticationToken == "" {
			authenticationToken = httpRequest.URL.Query().Get("token")
		}

		if authenticationToken == "" {
			http.Error(responseWriter, "Authentication required: no bearer token provided", http.StatusUnauthorized)
			return
		}

		// 3. Cryptographic Validation
		tokenClaims, validationError := auth.ValidateToken(authenticationToken)
		if validationError != nil {
			http.Error(responseWriter, "Session expired or invalid token", http.StatusUnauthorized)
			return
		}

		// 4. Session Integrity Check (Fingerprinting Verification)
		// This prevents person A from using person B's token even if they steal it,
		// because their IP or User-Agent profile will differ.
		requesterIP := httpRequest.RemoteAddr
		requesterUserAgent := httpRequest.Header.Get("User-Agent")
		currentClientFingerprint := auth.GenerateClientFingerprint(requesterIP, requesterUserAgent)

		if tokenClaims.ClientFingerprint != currentClientFingerprint {
			// SECURITY BREACH DETECTED: Token used from unauthorized device/browser
			http.Error(responseWriter, "Security Violation: Access denied due to unauthorized device signature", http.StatusForbidden)
			return
		}

		// 5. Context Enrichment
		requestContext := context.WithValue(httpRequest.Context(), UserIDKey, tokenClaims.UserID)
		requestContext = context.WithValue(requestContext, UserRoleKey, tokenClaims.Role)
		
		next.ServeHTTP(responseWriter, httpRequest.WithContext(requestContext))
	})
}
