package handlers

import (
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
	var u usermodel.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hashed, _ := auth.HashPassword(u.Password)
	u.Password = hashed

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

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	u, err := h.Store.GetByEmail(input.Email)
	if err != nil || !auth.CheckPasswordHash(input.Password, u.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(u.ID)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
