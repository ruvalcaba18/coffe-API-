package handlers

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/internal/auth"
	usermodel "coffeebase-api/internal/models/user"
	userstore "coffeebase-api/internal/store/user"
	"encoding/json"
	webServer "net/http"
)

/**
 * AuthHandler handles user registration and authentication.
 * Refactored to eliminate all shorthands and follow strictly declarative naming.
 */
type AuthHandler struct {
	Store *userstore.Store
}

func (handler *AuthHandler) Register(responseWriter webServer.ResponseWriter, httpRequest *webServer.Request) {
	var registrationRequest dto.RegisterRequest
	decodingError := json.NewDecoder(httpRequest.Body).Decode(&registrationRequest)
	if decodingError != nil {
		webServer.Error(responseWriter, "Invalid request body", webServer.StatusBadRequest)
		return
	}

	hashedPassword, _ := auth.HashPassword(registrationRequest.Password)
	
	userInstance := usermodel.User{
		Username: registrationRequest.Username,
		Email:    registrationRequest.Email,
		Password: hashedPassword,
		Language: registrationRequest.Language,
	}

	// Default language if not provided
	if userInstance.Language == "" {
		userInstance.Language = "es"
	}

	validLanguages := map[string]bool{"es": true, "en": true, "fr": true, "de": true, "gsw": true}
	if !validLanguages[userInstance.Language] {
		webServer.Error(responseWriter, "Invalid language", webServer.StatusBadRequest)
		return
	}

	creationError := handler.Store.Create(&userInstance)
	if creationError != nil {
		webServer.Error(responseWriter, "Error creating user", webServer.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(webServer.StatusCreated)
	json.NewEncoder(responseWriter).Encode(dto.MapUserToResponse(userInstance))
}

func (handler *AuthHandler) Login(responseWriter webServer.ResponseWriter, httpRequest *webServer.Request) {
	var loginRequest dto.LoginRequest
	decodingError := json.NewDecoder(httpRequest.Body).Decode(&loginRequest)
	if decodingError != nil {
		webServer.Error(responseWriter, "Invalid request body", webServer.StatusBadRequest)
		return
	}

	userInstance, fetchError := handler.Store.GetByEmail(loginRequest.Email)
	passwordMatch := auth.CheckPasswordHash(loginRequest.Password, userInstance.Password)

	if fetchError != nil || !passwordMatch {
		webServer.Error(responseWriter, "Invalid credentials", webServer.StatusUnauthorized)
		return
	}

	requesterIP := httpRequest.RemoteAddr
	requesterUserAgent := httpRequest.Header.Get("User-Agent")

	authenticationToken, tokenGenerationError := auth.GenerateToken(
		userInstance.ID, 
		userInstance.Role, 
		requesterIP, 
		requesterUserAgent,
	)
	if tokenGenerationError != nil {
		webServer.Error(responseWriter, "Error generating secure session", webServer.StatusInternalServerError)
		return
	}

	// Set secure HTTP-only cookie to prevent XSS theft
	authCookie := &webServer.Cookie{
		Name:     "auth-token",
		Value:    authenticationToken,
		Path:     "/",
		MaxAge:   7200, // 2 hours
		HttpOnly: true,
		Secure:   false, // Set to true in production over HTTPS
		SameSite: webServer.SameSiteLaxMode,
	}
	webServer.SetCookie(responseWriter, authCookie)

	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(map[string]string{"token": authenticationToken})
}
