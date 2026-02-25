package handlers

import (
	"coffeebase-api/internal/middleware"
	"coffeebase-api/internal/store/favorite"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type FavoriteHandler struct {
	Store *favorite.Store
}

func (h *FavoriteHandler) Add(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var input struct {
		ProductID int `json:"product_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Store.Add(userID, input.ProductID); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Product added to favorites"})
}

func (h *FavoriteHandler) Remove(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	productID, _ := strconv.Atoi(chi.URLParam(r, "id"))

	if err := h.Store.Remove(userID, productID); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Product removed from favorites"})
}

func (h *FavoriteHandler) GetUserFavorites(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	favorites, err := h.Store.GetUserFavorites(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favorites)
}
