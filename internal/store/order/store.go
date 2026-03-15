package order

import (
	ordermodel "coffeebase-api/internal/models/order"
	"context"
	"database/sql"
)

type Store interface {
	Create(requestContext context.Context, orderInstance *ordermodel.Order) error
	CreateWithTx(requestContext context.Context, databaseTransaction *sql.Tx, orderInstance *ordermodel.Order) error
	GetByID(requestContext context.Context, orderID string) (ordermodel.Order, error)
	GetByUserID(requestContext context.Context, userID int) ([]ordermodel.Order, error)
	GetLatestByUserID(requestContext context.Context, userID int) (ordermodel.Order, error)
	GetPickupsByUserID(requestContext context.Context, userID int) ([]ordermodel.Order, error)
	GetAll(requestContext context.Context) ([]ordermodel.Order, error)
	UpdateStatus(requestContext context.Context, orderID string, status string) error
	GetDashboardStats(requestContext context.Context) (DashboardStats, error)
}

type postgresStore struct {
	databaseConnection *sql.DB
}

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

// --- Public ---

func NewStore(databaseConnection *sql.DB) Store {
	return &postgresStore{databaseConnection: databaseConnection}
}
