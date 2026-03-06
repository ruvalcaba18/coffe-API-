package admin

import (
	"coffeebase-api/internal/store/order"
	userstore "coffeebase-api/internal/store/user"
	"encoding/json"
	"net/http"
)

type DashboardHandler struct {
	OrderStore *order.Store
	UserStore  *userstore.Store
}

func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	orderStats, err := h.OrderStore.GetDashboardStats()
	if err != nil {
		http.Error(w, "Failed to fetch order statistics", http.StatusInternalServerError)
		return
	}

	// Fetch user count
	userCount, err := h.UserStore.GetTotalCount()
	if err != nil {
		// Non-critical, log and continue
		userCount = 0
	}

	response := map[string]interface{}{
		"orders": orderStats,
		"users": map[string]interface{}{
			"total_count": userCount,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
