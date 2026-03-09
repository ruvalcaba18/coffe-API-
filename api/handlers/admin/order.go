package admin

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/internal/notifications"
	orderstore "coffeebase-api/internal/store/order"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	OrderStore      *orderstore.Store
	NotificationHub *notifications.Hub
}

func (orderHandler *OrderHandler) GetAll(responseWriter http.ResponseWriter, request *http.Request) {
	orderList, fetchError := orderHandler.OrderStore.GetAll()
	if fetchError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(dto.MapOrdersToResponse(orderList))
}

func (orderHandler *OrderHandler) UpdateStatus(responseWriter http.ResponseWriter, request *http.Request) {
	orderIdentifier := chi.URLParam(request, "id")
	var statusInput struct {
		Status string `json:"status"`
	}
	if decodeError := json.NewDecoder(request.Body).Decode(&statusInput); decodeError != nil {
		http.Error(responseWriter, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Get the order to know which user to notify
	orderInstance, fetchError := orderHandler.OrderStore.GetByID(orderIdentifier)
	if fetchError != nil {
		http.Error(responseWriter, "Order not found", http.StatusNotFound)
		return
	}

	// 2. Update in DB
	if updateError := orderHandler.OrderStore.UpdateStatus(orderIdentifier, statusInput.Status); updateError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 3. Send real-time notification via WebSocket
	orderHandler.NotificationHub.SendToUser(orderInstance.UserID, map[string]interface{}{
		"type":     "ORDER_UPDATE",
		"order_id": orderIdentifier,
		"status":   statusInput.Status,
		"message":  "¡Tu pedido ha cambiado de estado!",
	})

	responseWriter.WriteHeader(http.StatusOK)
	json.NewEncoder(responseWriter).Encode(map[string]string{"message": "Order status updated and notified"})
}
