package handlers

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/api/response"
	"coffeebase-api/internal/apperrors"
	"coffeebase-api/internal/middleware"
	"coffeebase-api/internal/store/favorite"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type FavoriteHandler struct {
	favoriteStore favorite.Store
}

// --- Public ---

func NewFavoriteHandler(favoriteStore favorite.Store) *FavoriteHandler {
	return &FavoriteHandler{
		favoriteStore: favoriteStore,
	}
}

func (favoriteHandler *FavoriteHandler) Add(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userID := httpRequest.Context().Value(middleware.UserIDKey).(int)

	var request dto.FavoriteRequest
	if error := response.DecodeJSON(httpRequest, &request); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	if error := favoriteHandler.favoriteStore.Add(httpRequest.Context(), userID, request.ProductID); error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, map[string]string{"message": "Product added to favorites"})
}

func (favoriteHandler *FavoriteHandler) Remove(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userID := httpRequest.Context().Value(middleware.UserIDKey).(int)
	productIDString := chi.URLParam(httpRequest, "id")
	productID, conversionError := strconv.Atoi(productIDString)
	if conversionError != nil {
		response.SendError(responseWriter, apperrors.ErrInvalidID)
		return
	}

	if error := favoriteHandler.favoriteStore.Remove(httpRequest.Context(), userID, productID); error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, map[string]string{"message": "Product removed from favorites"})
}

func (favoriteHandler *FavoriteHandler) GetUserFavorites(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userID := httpRequest.Context().Value(middleware.UserIDKey).(int)

	favoriteList, error := favoriteHandler.favoriteStore.GetUserFavorites(httpRequest.Context(), userID)
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, dto.MapProductsToResponse(favoriteList))
}
