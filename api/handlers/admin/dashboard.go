package admin

import (
	"coffeebase-api/api/dto"
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

	totalUserCount, userCountError := dashboardHandler.UserStore.GetTotalCount()
	if userCountError != nil {
		totalUserCount = 0
	}

	dashboardResponse := dto.DashboardStatsDTO{
		Orders: dto.DashboardOrderStatsDTO{
			TotalOrders:       orderStatistics.TotalOrders,
			TotalRevenue:      orderStatistics.TotalRevenue,
			AverageOrderValue: orderStatistics.AverageOrderValue,
			PendingOrders:     orderStatistics.PendingOrders,
			TakeoutOrders:     orderStatistics.TakeoutOrders,
			SalesHistory:      orderStatistics.SalesHistory,
		},
		Users: map[string]interface{}{
			"total_count": totalUserCount,
		},
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(dashboardResponse)
}
