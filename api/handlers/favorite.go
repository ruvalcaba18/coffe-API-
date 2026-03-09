package handlers

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/internal/middleware"
	"coffeebase-api/internal/store/favorite"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type FavoriteHandler struct {
	FavoriteStore *favorite.Store
}

func (favoriteHandler *FavoriteHandler) Add(responseWriter http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.UserIDKey).(int)

	var addInput dto.FavoriteRequest
	if decodeError := json.NewDecoder(request.Body).Decode(&addInput); decodeError != nil {
		http.Error(responseWriter, "Invalid request body", http.StatusBadRequest)
		return
	}

	if additionError := favoriteHandler.FavoriteStore.Add(userID, addInput.ProductID); additionError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseWriter.WriteHeader(http.StatusOK)
	json.NewEncoder(responseWriter).Encode(map[string]string{"message": "Product added to favorites"})
}

func (favoriteHandler *FavoriteHandler) Remove(responseWriter http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.UserIDKey).(int)
	productIdentifier, _ := strconv.Atoi(chi.URLParam(request, "id"))

	if removalError := favoriteHandler.FavoriteStore.Remove(userID, productIdentifier); removalError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseWriter.WriteHeader(http.StatusOK)
	json.NewEncoder(responseWriter).Encode(map[string]string{"message": "Product removed from favorites"})
}

func (favoriteHandler *FavoriteHandler) GetUserFavorites(responseWriter http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.UserIDKey).(int)

	favoriteList, fetchError := favoriteHandler.FavoriteStore.GetUserFavorites(userID)
	if fetchError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(dto.MapProductsToResponse(favoriteList))
}
