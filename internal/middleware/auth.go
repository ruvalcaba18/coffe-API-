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

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		var authenticationToken string

		authorizationHeader := httpRequest.Header.Get("Authorization")
		if authorizationHeader != "" {
			headerParts := strings.Split(authorizationHeader, " ")
			if len(headerParts) == 2 && headerParts[0] == "Bearer" {
				authenticationToken = headerParts[1]
			}
		}

		if authenticationToken == "" {
			authenticationToken = httpRequest.URL.Query().Get("token")
		}

		if authenticationToken == "" {
			http.Error(responseWriter, "Authentication required: no bearer token provided", http.StatusUnauthorized)
			return
		}

		tokenClaims, validationError := auth.ValidateToken(authenticationToken)
		if validationError != nil {
			http.Error(responseWriter, "Session expired or invalid token", http.StatusUnauthorized)
			return
		}

		requesterIP := httpRequest.RemoteAddr
		requesterUserAgent := httpRequest.Header.Get("User-Agent")
		currentClientFingerprint := auth.GenerateClientFingerprint(requesterIP, requesterUserAgent)

		if tokenClaims.ClientFingerprint != currentClientFingerprint {
			http.Error(responseWriter, "Security Violation: Access denied due to unauthorized device signature", http.StatusForbidden)
			return
		}

		requestContext := context.WithValue(httpRequest.Context(), UserIDKey, tokenClaims.UserID)
		requestContext = context.WithValue(requestContext, UserRoleKey, tokenClaims.Role)
		
		next.ServeHTTP(responseWriter, httpRequest.WithContext(requestContext))
	})
}
