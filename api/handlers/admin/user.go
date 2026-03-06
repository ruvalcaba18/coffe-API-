package admin

import (
	"coffeebase-api/api/dto"
	userstore "coffeebase-api/internal/store/user"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	Store *userstore.Store
}

func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.Store.GetAll()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var resp []dto.UserResponse
	for _, u := range users {
		resp = append(resp, dto.UserResponse{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			Language:  u.Language,
			AvatarURL: u.AvatarURL,
			Role:      u.Role,
			CreatedAt: u.CreatedAt,
		})
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Store.UpdateRole(id, req.Role); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
