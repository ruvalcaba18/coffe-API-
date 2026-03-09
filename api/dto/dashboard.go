package dto

// DashboardOrderStatsDTO represents order-related metrics
type DashboardOrderStatsDTO struct {
	TotalOrders       int     `json:"total_orders"`
	TotalRevenue      float64 `json:"total_revenue"`
	AverageOrderValue float64 `json:"avg_order_value"`
	PendingOrders     int     `json:"pending_orders"`
	TakeoutOrders     int     `json:"takeout_orders"`
}

// DashboardStatsDTO represents the statistics returned for the admin dashboard
type DashboardStatsDTO struct {
	Orders DashboardOrderStatsDTO `json:"orders"`
	Users  map[string]interface{} `json:"users"`
}
