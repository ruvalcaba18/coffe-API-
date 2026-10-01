package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// AuditMiddleware loggea eventos de seguridad importantes (OWASP A09).
// Se aplica selectivamente a rutas sensibles: login, logout, refresh, attendance.
func AuditMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		start := time.Now()
		wrapped := &auditResponseWriter{ResponseWriter: responseWriter, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, httpRequest)

		// Log de eventos de seguridad
		path := httpRequest.URL.Path
		method := httpRequest.Method
		status := wrapped.statusCode
		ip := extractClientIP(httpRequest)
		duration := time.Since(start)

		if isSecuritySensitivePath(path) {
			level := slog.LevelInfo
			if status >= 400 {
				level = slog.LevelWarn
			}
			if status >= 500 {
				level = slog.LevelError
			}

			slog.Log(httpRequest.Context(), level, "[AUDIT] Security event",
				"method", method,
				"path", path,
				"status", status,
				"ip", ip,
				"duration_ms", duration.Milliseconds(),
			)
		}
	})
}

// --- Private ---

type auditResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *auditResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func isSecuritySensitivePath(path string) bool {
	sensitivePaths := []string{
		"/api/v1/tokens",
		"/api/v1/logout",
		"/api/v1/attendance",
		"/api/v1/admin/attendance",
	}
	for _, sensitive := range sensitivePaths {
		if strings.HasPrefix(path, sensitive) {
			return true
		}
	}
	return false
}

func extractClientIP(httpRequest *http.Request) string {
	ip := httpRequest.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = httpRequest.RemoteAddr
	}
	return strings.Split(strings.Split(ip, ",")[0], ":")[0]
}
