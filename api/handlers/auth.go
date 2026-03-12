package handlers

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/internal/auth"
	usermodel "coffeebase-api/internal/models/user"
	"coffeebase-api/internal/notifications"
	userstore "coffeebase-api/internal/store/user"
	"encoding/json"
	"log"
	webServer "net/http"
	stringManipulation "strings"
)

type AuthHandler struct {
	UserStore       *userstore.Store
	NotificationHub *notifications.Hub
}

func (authHandler *AuthHandler) Register(responseWriter webServer.ResponseWriter, httpRequest *webServer.Request) {
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

	if userInstance.Language == "" {
		userInstance.Language = "es"
	}

	validLanguages := map[string]bool{"es": true, "en": true, "fr": true, "de": true, "gsw": true}
	if !validLanguages[userInstance.Language] {
		webServer.Error(responseWriter, "Invalid language", webServer.StatusBadRequest)
		return
	}

	creationError := authHandler.UserStore.Create(&userInstance)
	if creationError != nil {
		webServer.Error(responseWriter, "Error creating user", webServer.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(webServer.StatusCreated)
	userResponse := dto.MapUserToResponse(userInstance)
	
	if authHandler.NotificationHub != nil {
		authHandler.NotificationHub.Broadcast(map[string]interface{}{
			"type": "new_user",
			"user": userResponse,
		})
	}

	json.NewEncoder(responseWriter).Encode(userResponse)
}

func (authHandler *AuthHandler) Login(responseWriter webServer.ResponseWriter, httpRequest *webServer.Request) {
	var loginRequest dto.LoginRequest
	decodingError := json.NewDecoder(httpRequest.Body).Decode(&loginRequest)
	if decodingError != nil {
		webServer.Error(responseWriter, "Invalid request body", webServer.StatusBadRequest)
		return
	}

	loginRequest.Email = stringManipulation.ToLower(stringManipulation.TrimSpace(loginRequest.Email))
	userInstance, fetchError := authHandler.UserStore.GetByEmail(loginRequest.Email)
	passwordMatch := auth.CheckPasswordHash(loginRequest.Password, userInstance.Password)

	if fetchError != nil {
		log.Printf("Login failed: User not found with email %s", loginRequest.Email)
		webServer.Error(responseWriter, "Invalid credentials", webServer.StatusUnauthorized)
		return
	}

	if !passwordMatch {
		log.Printf("Login failed: Password mismatch for email %s", loginRequest.Email)
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

	authCookie := &webServer.Cookie{
		Name:     "auth-token",
		Value:    authenticationToken,
		Path:     "/",
		MaxAge:   7200, 
		HttpOnly: true,
		Secure:   false, 
		SameSite: webServer.SameSiteLaxMode,
	}
	webServer.SetCookie(responseWriter, authCookie)

	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(map[string]string{"token": authenticationToken})
}
