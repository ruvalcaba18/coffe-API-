package handlers

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/api/response"
	"coffeebase-api/internal/apperrors"
	"coffeebase-api/internal/auth"
	usermodel "coffeebase-api/internal/models/user"
	tokenmodel "coffeebase-api/internal/models/token"
	"coffeebase-api/internal/notifications"
	tokenstore "coffeebase-api/internal/store/token"
	"coffeebase-api/internal/validation"
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	accessTokenExpiresIn  = 7200 // 2 horas en segundos
	refreshTokenExpiresIn = 7 * 24 * time.Hour
)

type AuthHandler struct {
	userStore       UserAuthStore
	tokenStore      tokenstore.Store
	notificationHub *notifications.Hub
}

type UserAuthStore interface {
	Create(requestContext context.Context, user *usermodel.User) error
	GetByEmail(requestContext context.Context, email string) (usermodel.User, error)
}

// --- Public ---

func NewAuthHandler(userStore UserAuthStore, tokenStore tokenstore.Store, notificationHub *notifications.Hub) *AuthHandler {
	return &AuthHandler{
		userStore:       userStore,
		tokenStore:      tokenStore,
		notificationHub: notificationHub,
	}
}

// Register crea un nuevo usuario (OWASP A03: validación y sanitización de inputs).
func (h *AuthHandler) Register(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	var request dto.RegisterRequest
	if error := response.DecodeJSON(httpRequest, &request); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	userInstance, error := h.buildUserFromRegistration(request)
	if error != nil {
		response.SendError(responseWriter, error)
		return
	}

	if error := h.userStore.Create(httpRequest.Context(), userInstance); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	h.finalizeRegistration(responseWriter, userInstance)
}

// Login autentica al usuario y devuelve el par {access_token, refresh_token}.
func (h *AuthHandler) Login(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	var request dto.LoginRequest
	if error := response.DecodeJSON(httpRequest, &request); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	userInstance, error := h.authenticateUser(httpRequest, request)
	if error != nil {
		response.SendError(responseWriter, error)
		return
	}

	tokenPair, error := h.createTokenPair(httpRequest, userInstance)
	if error != nil {
		response.SendError(responseWriter, error)
		return
	}

	// OWASP A04: Cookie HttpOnly para el refresh token (doble canal)
	h.setRefreshCookie(responseWriter, tokenPair.RefreshToken)
	// OWASP A04: Cookie HttpOnly para el access token
	h.setAuthCookie(responseWriter, tokenPair.AccessToken)

	response.SendJSON(responseWriter, http.StatusOK, tokenPair)
}

// Refresh rota el par de tokens: invalida el refresh viejo y emite uno nuevo.
// Si el refresh token ya fue usado → robo detectado → invalida toda la familia.
func (h *AuthHandler) Refresh(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	rawRefreshToken := h.extractRefreshToken(httpRequest)
	if rawRefreshToken == "" {
		response.SendError(responseWriter, apperrors.ErrRefreshTokenInvalid)
		return
	}

	tokenHash := auth.HashRefreshToken(rawRefreshToken)
	storedToken, err := h.tokenStore.GetByHash(httpRequest.Context(), tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			response.SendError(responseWriter, apperrors.ErrRefreshTokenInvalid)
		} else {
			response.SendError(responseWriter, apperrors.ErrInternalServerError)
		}
		return
	}

	// ¿Ya fue usado? → Posible robo → invalida toda la familia
	if storedToken.Used {
		slog.Warn("[SECURITY] Refresh token reuse detected — invalidating family",
			"family_id", storedToken.FamilyID,
			"user_id", storedToken.UserID,
		)
		_ = h.tokenStore.InvalidateFamily(httpRequest.Context(), storedToken.FamilyID)
		response.SendError(responseWriter, apperrors.ErrRefreshTokenRevoked)
		return
	}

	// ¿Expirado?
	if time.Now().After(storedToken.ExpiresAt) {
		response.SendError(responseWriter, apperrors.ErrRefreshTokenInvalid)
		return
	}

	// Marcar el token viejo como usado (rotación)
	if err := h.tokenStore.MarkUsed(httpRequest.Context(), storedToken.ID); err != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	// Generar nuevo par con el mismo family_id
	ip := extractIP(httpRequest)
	ua := httpRequest.Header.Get("User-Agent")

	accessToken, newRawRefresh, err := auth.GenerateTokenPair(
		storedToken.UserID, "", ip, ua,
	)
	if err != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	// Para incluir el rol en el nuevo access token necesitamos recuperarlo del token viejo
	// El access token lleva el rol en los claims — lo re-extraemos del token enviado en header
	if accessTokenFromHeader := extractAccessToken(httpRequest); accessTokenFromHeader != "" {
		if claims, claimsErr := auth.ValidateToken(accessTokenFromHeader); claimsErr == nil {
			accessToken, _, err = auth.GenerateTokenPair(storedToken.UserID, claims.Role, ip, ua)
			if err != nil {
				response.SendError(responseWriter, apperrors.ErrInternalServerError)
				return
			}
		}
	}

	newRefreshHash := auth.HashRefreshToken(newRawRefresh)
	newToken := &tokenmodel.RefreshToken{
		ID:        uuid.New(),
		FamilyID:  storedToken.FamilyID, // mismo family_id
		UserID:    storedToken.UserID,
		TokenHash: newRefreshHash,
		ExpiresAt: time.Now().Add(refreshTokenExpiresIn),
	}

	if err := h.tokenStore.Create(httpRequest.Context(), newToken); err != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	tokenPair := dto.TokenPairResponse{
		AccessToken:  accessToken,
		RefreshToken: newRawRefresh,
		ExpiresIn:    accessTokenExpiresIn,
	}

	h.setRefreshCookie(responseWriter, newRawRefresh)
	h.setAuthCookie(responseWriter, accessToken)
	response.SendJSON(responseWriter, http.StatusOK, tokenPair)
}

// Logout invalida el access token (Redis blacklist) y toda la familia de refresh tokens.
func (h *AuthHandler) Logout(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	// Revocar el refresh token y su familia
	rawRefreshToken := h.extractRefreshToken(httpRequest)
	if rawRefreshToken != "" {
		tokenHash := auth.HashRefreshToken(rawRefreshToken)
		if stored, err := h.tokenStore.GetByHash(httpRequest.Context(), tokenHash); err == nil {
			_ = h.tokenStore.InvalidateFamily(httpRequest.Context(), stored.FamilyID)
		}
	}

	// Limpiar cookies
	h.clearAuthCookies(responseWriter)
	responseWriter.WriteHeader(http.StatusNoContent)
}

// --- Private ---

func (h *AuthHandler) buildUserFromRegistration(request dto.RegisterRequest) (*usermodel.User, error) {
	cleanEmail, error := validation.Email(request.Email)
	if error != nil {
		return nil, apperrors.ErrInvalidRequest
	}

	cleanUsername, error := validation.Username(request.Username)
	if error != nil {
		return nil, apperrors.ErrInvalidRequest
	}

	if error := validation.Password(request.Password); error != nil {
		return nil, apperrors.ErrInvalidRequest
	}

	hashedPassword, error := auth.HashPassword(request.Password)
	if error != nil {
		return nil, apperrors.ErrInternalServerError
	}

	language := strings.ToLower(strings.TrimSpace(request.Language))
	if language == "" {
		language = "es"
	}
	if !isValidLanguage(language) {
		return nil, apperrors.ErrInvalidRequest
	}

	return &usermodel.User{
		Username: cleanUsername,
		Email:    cleanEmail,
		Password: hashedPassword,
		Language: language,
	}, nil
}

func (h *AuthHandler) finalizeRegistration(responseWriter http.ResponseWriter, user *usermodel.User) {
	userResponse := dto.MapUserToResponse(*user)
	if h.notificationHub != nil {
		h.notificationHub.Broadcast(map[string]interface{}{
			"type": "new_user",
			"user": userResponse,
		})
	}
	response.SendJSON(responseWriter, http.StatusCreated, userResponse)
}

func (h *AuthHandler) authenticateUser(httpRequest *http.Request, request dto.LoginRequest) (usermodel.User, error) {
	email := strings.ToLower(strings.TrimSpace(request.Email))
	userInstance, error := h.userStore.GetByEmail(httpRequest.Context(), email)
	if error != nil {
		return usermodel.User{}, apperrors.ErrUnauthorized
	}
	if !auth.CheckPasswordHash(request.Password, userInstance.Password) {
		return usermodel.User{}, apperrors.ErrUnauthorized
	}
	return userInstance, nil
}

func (h *AuthHandler) createTokenPair(httpRequest *http.Request, user usermodel.User) (dto.TokenPairResponse, error) {
	ip := extractIP(httpRequest)
	ua := httpRequest.Header.Get("User-Agent")

	accessToken, rawRefreshToken, err := auth.GenerateTokenPair(user.ID, string(user.Role), ip, ua)
	if err != nil {
		return dto.TokenPairResponse{}, apperrors.ErrInternalServerError
	}

	refreshHash := auth.HashRefreshToken(rawRefreshToken)
	refreshToken := &tokenmodel.RefreshToken{
		ID:        uuid.New(),
		FamilyID:  uuid.New(), // nueva familia por cada login
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: time.Now().Add(refreshTokenExpiresIn),
	}

	if err := h.tokenStore.Create(httpRequest.Context(), refreshToken); err != nil {
		return dto.TokenPairResponse{}, apperrors.ErrInternalServerError
	}

	return dto.TokenPairResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresIn:    accessTokenExpiresIn,
	}, nil
}

func (h *AuthHandler) extractRefreshToken(httpRequest *http.Request) string {
	// Prioridad 1: cookie HttpOnly (más seguro)
	if cookie, err := httpRequest.Cookie("refresh-token"); err == nil {
		return cookie.Value
	}
	// Prioridad 2: body JSON
	var req dto.RefreshRequest
	if err := response.DecodeJSON(httpRequest, &req); err == nil && req.RefreshToken != "" {
		return req.RefreshToken
	}
	return ""
}

func extractAccessToken(httpRequest *http.Request) string {
	authHeader := httpRequest.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}
	return ""
}

func (h *AuthHandler) setAuthCookie(responseWriter http.ResponseWriter, token string) {
	secureCookie := os.Getenv("ENV") == "production"
	http.SetCookie(responseWriter, &http.Cookie{
		Name:     "auth-token",
		Value:    token,
		Path:     "/",
		MaxAge:   accessTokenExpiresIn,
		HttpOnly: true,
		Secure:   secureCookie,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *AuthHandler) setRefreshCookie(responseWriter http.ResponseWriter, rawRefreshToken string) {
	secureCookie := os.Getenv("ENV") == "production"
	http.SetCookie(responseWriter, &http.Cookie{
		Name:     "refresh-token",
		Value:    rawRefreshToken,
		Path:     "/api/v1/tokens/refresh",  // solo accesible desde el endpoint de refresh
		MaxAge:   int(refreshTokenExpiresIn.Seconds()),
		HttpOnly: true,
		Secure:   secureCookie,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *AuthHandler) clearAuthCookies(responseWriter http.ResponseWriter) {
	http.SetCookie(responseWriter, &http.Cookie{Name: "auth-token", MaxAge: -1, Path: "/"})
	http.SetCookie(responseWriter, &http.Cookie{Name: "refresh-token", MaxAge: -1, Path: "/api/v1/tokens/refresh"})
}

func extractIP(httpRequest *http.Request) string {
	ip := httpRequest.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = httpRequest.RemoteAddr
	}
	return strings.Split(strings.Split(ip, ",")[0], ":")[0]
}

func isValidLanguage(lang string) bool {
	validLanguages := map[string]bool{"es": true, "en": true, "fr": true, "de": true, "gsw": true}
	return validLanguages[lang]
}
