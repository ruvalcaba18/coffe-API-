package handlers

import (
	"encoding/json"
	"net/http"
)

type BillingHandler struct{}

func (h *BillingHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"balance": 150.50, // Mock balance
	})
}

func (h *BillingHandler) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	methods := []map[string]interface{}{
		{
			"id":     1,
			"last4":  "4242",
			"expiry": "12/26",
			"brand":  "Visa",
			"holder": "Ian Admin",
		},
		{
			"id":     2,
			"last4":  "5555",
			"expiry": "09/25",
			"brand":  "Mastercard",
			"holder": "Ian Admin",
		},
	}
	json.NewEncoder(w).Encode(methods)
}
