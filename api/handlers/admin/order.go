package admin

import (
	"coffeebase-api/internal/notifications"
	orderstore "coffeebase-api/internal/store/order"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	Store *orderstore.Store
	Hub   *notifications.Hub
}

func (h *OrderHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	orders, err := h.Store.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Get the order to know which user to notify
	o, err := h.Store.GetByID(id)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// 2. Update in DB
	if err := h.Store.UpdateStatus(id, input.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Send real-time notification via WebSocket
	h.Hub.SendToUser(o.UserID, map[string]interface{}{
		"type":     "ORDER_UPDATE",
		"order_id": id,
		"status":   input.Status,
		"message":  "¡Tu pedido ha cambiado de estado!",
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Order status updated and notified"})
}
