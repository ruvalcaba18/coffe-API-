package admin

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/api/response"
	"coffeebase-api/internal/apperrors"
	"coffeebase-api/internal/notifications"
	orderstore "coffeebase-api/internal/store/order"
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	orderStore      orderstore.Store
	notificationHub *notifications.Hub
}

// --- Public ---

func NewOrderHandler(orderStore orderstore.Store, notificationHub *notifications.Hub) *OrderHandler {
	return &OrderHandler{
		orderStore:      orderStore,
		notificationHub: notificationHub,
	}
}

func (orderHandler *OrderHandler) GetAll(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	orderList, error := orderHandler.orderStore.GetAll(httpRequest.Context())
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, dto.MapOrdersToResponse(orderList))
}

func (orderHandler *OrderHandler) UpdateStatus(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	orderIdentifier := chi.URLParam(httpRequest, "id")
	
	var request struct {
		Status string `json:"status"`
	}
	if error := response.DecodeJSON(httpRequest, &request); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	orderInstance, error := orderHandler.orderStore.GetByID(httpRequest.Context(), orderIdentifier)
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	if error := orderHandler.updateAndNotifyStatus(httpRequest.Context(), orderIdentifier, request.Status, orderInstance.UserID); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, map[string]string{"message": "Order status updated and notified"})
}

// --- Private ---

func (orderHandler *OrderHandler) updateAndNotifyStatus(requestContext context.Context, orderID string, status string, userID int) error {
	if error := orderHandler.orderStore.UpdateStatus(requestContext, orderID, status); error != nil {
		return apperrors.ErrInternalServerError
	}

	if orderHandler.notificationHub != nil {
		orderHandler.notificationHub.SendToUser(userID, map[string]interface{}{
			"type":     "ORDER_UPDATE",
			"order_id": orderID,
			"status":   status,
			"message":  "¡Tu pedido ha cambiado de estado!",
		})
	}
	
	return nil
}
