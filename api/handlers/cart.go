package handlers

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/api/response"
	"coffeebase-api/internal/apperrors"
	"coffeebase-api/internal/middleware"
	"coffeebase-api/internal/store/cart"
	"net/http"
)

type CartHandler struct {
	cartStore cart.Store
}

// --- Public ---

func NewCartHandler(cartStore cart.Store) *CartHandler {
	return &CartHandler{
		cartStore: cartStore,
	}
}

func (cartHandler *CartHandler) GetCart(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userID := httpRequest.Context().Value(middleware.UserIDKey).(int)
	userCart, error := cartHandler.cartStore.GetCart(httpRequest.Context(), userID)
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}
	
	response.SendJSON(responseWriter, http.StatusOK, dto.MapCartToResponse(*userCart))
}

func (cartHandler *CartHandler) UpdateItem(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userID := httpRequest.Context().Value(middleware.UserIDKey).(int)
	
	var request dto.CartUpdateRequest
	if error := response.DecodeJSON(httpRequest, &request); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	if error := cartHandler.cartStore.UpdateItem(httpRequest.Context(), userID, request.ProductID, request.Quantity); error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, map[string]string{"message": "Cart updated"})
}
