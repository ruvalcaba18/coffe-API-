package admin

import (
	"coffeebase-api/internal/models/coupon"
	couponstore "coffeebase-api/internal/store/coupon"
	"encoding/json"
	"net/http"
)

type CouponHandler struct {
	Store *couponstore.Store
}

func (h *CouponHandler) Create(w http.ResponseWriter, r *http.Request) {
	var c coupon.Coupon
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Store.Create(&c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

func (h *CouponHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	coupons, err := h.Store.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(coupons)
}
