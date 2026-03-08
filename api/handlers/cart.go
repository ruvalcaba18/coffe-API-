package handlers

import (
	"coffeebase-api/internal/middleware"
	"coffeebase-api/internal/store/cart"
	"encoding/json"
	"net/http"
)

type CartHandler struct {
	CartStore *cart.Store
}

func (cartHandler *CartHandler) GetCart(responseWriter http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.UserIDKey).(int)
	userCart, fetchError := cartHandler.CartStore.GetCart(userID)
	if fetchError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}
	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(userCart)
}

func (cartHandler *CartHandler) UpdateItem(responseWriter http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.UserIDKey).(int)
	var cartUpdateInput struct {
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	}
	if decodeError := json.NewDecoder(request.Body).Decode(&cartUpdateInput); decodeError != nil {
		http.Error(responseWriter, "Invalid request", http.StatusBadRequest)
		return
	}

	if updateError := cartHandler.CartStore.UpdateItem(userID, cartUpdateInput.ProductID, cartUpdateInput.Quantity); updateError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseWriter.WriteHeader(http.StatusOK)
	json.NewEncoder(responseWriter).Encode(map[string]string{"message": "Cart updated"})
}
