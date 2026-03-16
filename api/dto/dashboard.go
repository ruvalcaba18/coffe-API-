package dto

type DailySaleDTO struct {
	Date        string  `json:"date"`
	Total       float64 `json:"total"`
	PickupCount int     `json:"pickup_count"`
}

type DashboardOrderStatsDTO struct {
	TotalOrders       int           `json:"total_orders"`
	TotalRevenue      float64       `json:"total_revenue"`
	AverageOrderValue float64       `json:"avg_order_value"`
	PendingOrders     int           `json:"pending_orders"`
	TakeoutOrders     int           `json:"takeout_orders"`
	SalesHistory      []DailySaleDTO `json:"sales_history"`
}

type DashboardStatsDTO struct {
	Orders  DashboardOrderStatsDTO `json:"orders"`
	Users   map[string]interface{} `json:"users"`
	Coupons map[string]interface{} `json:"coupons"`
}
