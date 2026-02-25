package handlers

import (
	"coffeebase-api/internal/middleware"
	ordermodel "coffeebase-api/internal/models/order"
	productmodel "coffeebase-api/internal/models/product"
	"context"
	"encoding/json"
	"net/http"
)

type OrderRepository interface {
	GetByUserID(userID int) ([]ordermodel.Order, error)
	Create(o *ordermodel.Order) error
}

type OrderCheckoutService interface {
	Checkout(ctx context.Context, userID int, couponCode string) (*ordermodel.Order, error)
}

type ProductReader interface {
	GetByID(id int) (productmodel.Product, error)
}

type OrderHandler struct {
	Store        OrderRepository
	ProductStore ProductReader
	Service      OrderCheckoutService
}

func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var input struct {
		CouponCode string `json:"coupon_code"`
	}
	json.NewDecoder(r.Body).Decode(&input)

	o, err := h.Service.Checkout(r.Context(), userID, input.CouponCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(o)
}

func (h *OrderHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	orders, err := h.Store.GetByUserID(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	var items []ordermodel.OrderItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var total float64
	for _, item := range items {
		p, _ := h.ProductStore.GetByID(item.ProductID)
		total += p.Price * float64(item.Quantity)
	}

	o := &ordermodel.Order{
		UserID: userID,
		Items:  items,
		Total:  total,
	}

	if err := h.Store.Create(o); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(o)
}
