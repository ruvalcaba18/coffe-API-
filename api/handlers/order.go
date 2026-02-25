package handlers

import (
	"coffeebase-api/internal/middleware"
	ordermodel "coffeebase-api/internal/models/order"
	orderservice "coffeebase-api/internal/service/order"
	cartstore "coffeebase-api/internal/store/cart"
	orderstore "coffeebase-api/internal/store/order"
	productstore "coffeebase-api/internal/store/product"
	"encoding/json"
	"net/http"
)

type OrderHandler struct {
	Store        *orderstore.Store
	ProductStore *productstore.Store
	CartStore    *cartstore.Store
	Service      *orderservice.Service
}

func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var input struct {
		CouponCode string `json:"coupon_code"`
	}
	json.NewDecoder(r.Body).Decode(&input)

	o, err := h.Service.Checkout(userID, input.CouponCode)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(o)
}
