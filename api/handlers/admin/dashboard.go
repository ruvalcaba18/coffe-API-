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

func (dashboardHandler *DashboardHandler) GetStats(responseWriter http.ResponseWriter, request *http.Request) {
	orderStatistics, orderStatsError := dashboardHandler.OrderStore.GetDashboardStats()
	if orderStatsError != nil {
		http.Error(responseWriter, "Failed to fetch order statistics", http.StatusInternalServerError)
		return
	}

	// Fetch user count
	totalUserCount, userCountError := dashboardHandler.UserStore.GetTotalCount()
	if userCountError != nil {
		// Non-critical, log and continue
		totalUserCount = 0
	}

	dashboardResponse := map[string]interface{}{
		"orders": orderStatistics,
		"users": map[string]interface{}{
			"total_count": totalUserCount,
		},
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(dashboardResponse)
}
