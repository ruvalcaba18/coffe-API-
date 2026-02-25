package handlers

import (
	"coffeebase-api/internal/middleware"
	reviewmodel "coffeebase-api/internal/models/review"
	reviewstore "coffeebase-api/internal/store/review"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ReviewHandler struct {
	Store *reviewstore.Store
}

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	productID, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var rev reviewmodel.Review
	if err := json.NewDecoder(r.Body).Decode(&rev); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	rev.UserID = userID
	rev.ProductID = productID

	if err := h.Store.Create(&rev); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rev)
}

func (h *ReviewHandler) GetByProduct(w http.ResponseWriter, r *http.Request) {
	productID, _ := strconv.Atoi(chi.URLParam(r, "id"))
	
	reviews, err := h.Store.GetByProductID(productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviews)
}
