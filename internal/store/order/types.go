package order

type DailySale struct {
	Date        string  `json:"date"`
	Total       float64 `json:"total"`
	PickupCount int     `json:"pickup_count"`
}

type DashboardStats struct {
	TotalOrders       int         `json:"total_orders"`
	TotalRevenue      float64     `json:"total_revenue"`
	AverageOrderValue float64     `json:"avg_order_value"`
	PendingOrders     int         `json:"pending_orders"`
	TakeoutOrders     int         `json:"takeout_orders"`
	SalesHistory      []DailySale `json:"sales_history"`
}
