package handlers

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/api/response"
	"coffeebase-api/internal/apperrors"
	"coffeebase-api/internal/auth"
	usermodel "coffeebase-api/internal/models/user"
	"coffeebase-api/internal/notifications"
	"coffeebase-api/internal/validation"
	"context"
	"net/http"
	"os"
	"strings"
)

type AuthHandler struct {
	userStore       UserAuthStore
	notificationHub *notifications.Hub
}

type UserAuthStore interface {
	Create(requestContext context.Context, user *usermodel.User) error
	GetByEmail(requestContext context.Context, email string) (usermodel.User, error)
}

// --- Public ---

func NewAuthHandler(userStore UserAuthStore, notificationHub *notifications.Hub) *AuthHandler {
	return &AuthHandler{
		userStore:       userStore,
		notificationHub: notificationHub,
	}
}

func (authHandler *AuthHandler) Register(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	var request dto.RegisterRequest
	if error := response.DecodeJSON(httpRequest, &request); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	userInstance, error := authHandler.buildUserFromRegistration(request)
	if error != nil {
		response.SendError(responseWriter, error)
		return
	}

	if error := authHandler.userStore.Create(httpRequest.Context(), userInstance); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	authHandler.finalizeRegistration(responseWriter, userInstance)
}

func (authHandler *AuthHandler) Login(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	var request dto.LoginRequest
	if error := response.DecodeJSON(httpRequest, &request); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	userInstance, error := authHandler.authenticateUser(httpRequest, request)
	if error != nil {
		response.SendError(responseWriter, error)
		return
	}

	token, error := authHandler.generateUserSession(httpRequest, userInstance)
	if error != nil {
		response.SendError(responseWriter, error)
		return
	}

	authHandler.setAuthCookie(responseWriter, token)
	response.SendJSON(responseWriter, http.StatusOK, map[string]string{"token": token})
}

// --- Private ---

func (authHandler *AuthHandler) buildUserFromRegistration(request dto.RegisterRequest) (*usermodel.User, error) {
	// OWASP A03 - Validate and sanitize all inputs
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

func (authHandler *AuthHandler) finalizeRegistration(responseWriter http.ResponseWriter, user *usermodel.User) {
	userResponse := dto.MapUserToResponse(*user)
	
	if authHandler.notificationHub != nil {
		authHandler.notificationHub.Broadcast(map[string]interface{}{
			"type": "new_user",
			"user": userResponse,
		})
	}

	response.SendJSON(responseWriter, http.StatusCreated, userResponse)
}

func (authHandler *AuthHandler) authenticateUser(httpRequest *http.Request, request dto.LoginRequest) (usermodel.User, error) {
	email := strings.ToLower(strings.TrimSpace(request.Email))
	userInstance, error := authHandler.userStore.GetByEmail(httpRequest.Context(), email)
	if error != nil {
		return usermodel.User{}, apperrors.ErrUnauthorized
	}

	if !auth.CheckPasswordHash(request.Password, userInstance.Password) {
		return usermodel.User{}, apperrors.ErrUnauthorized
	}

	return userInstance, nil
}

func (authHandler *AuthHandler) generateUserSession(httpRequest *http.Request, user usermodel.User) (string, error) {
	token, error := auth.GenerateToken(user.ID, string(user.Role), httpRequest.RemoteAddr, httpRequest.Header.Get("User-Agent"))
	if error != nil {
		return "", apperrors.ErrInternalServerError
	}
	return token, nil
}

func (authHandler *AuthHandler) setAuthCookie(responseWriter http.ResponseWriter, token string) {

	secureCookie := os.Getenv("ENV") == "production"
	http.SetCookie(responseWriter, &http.Cookie{
		Name:     "auth-token",
		Value:    token,
		Path:     "/",
		MaxAge:   7200,
		HttpOnly: true,     
		Secure:   secureCookie, 
		SameSite: http.SameSiteStrictMode, 
	})
}

func isValidLanguage(lang string) bool {
	validLanguages := map[string]bool{"es": true, "en": true, "fr": true, "de": true, "gsw": true}
	return validLanguages[lang]
}
