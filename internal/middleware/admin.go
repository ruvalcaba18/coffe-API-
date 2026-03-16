package middleware

import (
	"coffeebase-api/api/response"
	"coffeebase-api/internal/models/user"
	"net/http"
)

// --- Public ---

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		roleValue, ok := httpRequest.Context().Value(UserRoleKey).(string)
		role := user.UserRole(roleValue)
		
		if !ok || !hasStaffAccess(role) {
			response.Forbidden(responseWriter, "Staff access required")
			return
		}
		
		next.ServeHTTP(responseWriter, httpRequest)
	})
}

func SuperAdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		roleValue, ok := httpRequest.Context().Value(UserRoleKey).(string)
		role := user.UserRole(roleValue)
		
		if !ok || role != user.RoleSuperAdmin {
			response.Forbidden(responseWriter, "Super Admin access required")
			return
		}
		
		next.ServeHTTP(responseWriter, httpRequest)
	})
}

// --- Private ---

func hasStaffAccess(role user.UserRole) bool {
	return role == user.RoleAdmin || role == user.RoleSuperAdmin || role == user.RoleBarista
}
