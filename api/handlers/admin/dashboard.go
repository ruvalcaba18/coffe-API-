package admin

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/api/response"
	"coffeebase-api/internal/apperrors"
	"coffeebase-api/internal/store/order"
	userstore "coffeebase-api/internal/store/user"
	"net/http"
)

type DashboardHandler struct {
	orderStore order.Store
	userStore  userstore.Store
}

// --- Public ---

func NewDashboardHandler(orderStore order.Store, userStore userstore.Store) *DashboardHandler {
	return &DashboardHandler{
		orderStore: orderStore,
		userStore:  userStore,
	}
}

func (dashboardHandler *DashboardHandler) GetStats(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	orderStatistics, error := dashboardHandler.orderStore.GetDashboardStats(httpRequest.Context())
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	totalUserCount, error := dashboardHandler.userStore.GetTotalCount(httpRequest.Context())
	if error != nil {
		totalUserCount = 0
	}

	dashboardResponse := dto.DashboardStatsDTO{
		Orders: dto.DashboardOrderStatsDTO{
			TotalOrders:       orderStatistics.TotalOrders,
			TotalRevenue:      orderStatistics.TotalRevenue,
			AverageOrderValue: orderStatistics.AverageOrderValue,
			PendingOrders:     orderStatistics.PendingOrders,
			TakeoutOrders:     orderStatistics.TakeoutOrders,
			SalesHistory:      []dto.DailySaleDTO{},
		},
		Users: map[string]interface{}{
			"total_count": totalUserCount,
		},
	}

	for _, sale := range orderStatistics.SalesHistory {
		dashboardResponse.Orders.SalesHistory = append(dashboardResponse.Orders.SalesHistory, dto.DailySaleDTO{
			Date:        sale.Date,
			Total:       sale.Total,
			PickupCount: sale.PickupCount,
		})
	}

	response.SendJSON(responseWriter, http.StatusOK, dashboardResponse)
}
