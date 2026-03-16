package admin

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/api/response"
	"coffeebase-api/internal/apperrors"
	couponstore "coffeebase-api/internal/store/coupon"
	"coffeebase-api/internal/store/order"
	userstore "coffeebase-api/internal/store/user"
	"net/http"
)

type DashboardHandler struct {
	orderStore  order.Store
	userStore   userstore.Store
	couponStore couponstore.Store
}

// --- Public ---

func NewDashboardHandler(orderStore order.Store, userStore userstore.Store, couponStore couponstore.Store) *DashboardHandler {
	return &DashboardHandler{
		orderStore:  orderStore,
		userStore:   userStore,
		couponStore: couponStore,
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

	couponList, error := dashboardHandler.couponStore.GetAll(httpRequest.Context())
	if error != nil {
		couponList = nil
	}

	couponDTOs := dto.MapCouponsToResponse(couponList)
	totalCouponsUsed := 0
	for _, c := range couponList {
		totalCouponsUsed += c.UsedCount
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
		Coupons: map[string]interface{}{
			"list":       couponDTOs,
			"total_used": totalCouponsUsed,
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
