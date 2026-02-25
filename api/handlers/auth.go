package handlers

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/internal/auth"
	usermodel "coffeebase-api/internal/models/user"
	userstore "coffeebase-api/internal/store/user"
	"encoding/json"
	"net/http"
)

type AuthHandler struct {
	Store *userstore.Store
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	hashed, _ := auth.HashPassword(req.Password)
	
	u := usermodel.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashed,
		Language: req.Language,
	}

	// Default language if not provided
	if u.Language == "" {
		u.Language = "es"
	}

	validLanguages := map[string]bool{"es": true, "en": true, "fr": true, "de": true, "gsw": true}
	if !validLanguages[u.Language] {
		http.Error(w, "Invalid language", http.StatusBadRequest)
		return
	}

	if err := h.Store.Create(&u); err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.MapUserToResponse(u))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	u, err := h.Store.GetByEmail(req.Email)
	if err != nil || !auth.CheckPasswordHash(req.Password, u.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(u.ID, u.Role)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
