package order

import (
	ordermodel "coffeebase-api/internal/models/order"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Store struct {
	databaseConnection *sql.DB
}

func NewStore(databaseConnection *sql.DB) *Store {
	return &Store{databaseConnection: databaseConnection}
}

func (orderStore *Store) Create(orderInstance *ordermodel.Order) error {
	databaseTransaction, transactionBeginError := orderStore.databaseConnection.Begin()
	if transactionBeginError != nil {
		return transactionBeginError
	}
	defer databaseTransaction.Rollback()

	if createError := orderStore.CreateWithTx(databaseTransaction, orderInstance); createError != nil {
		return createError
	}

	return databaseTransaction.Commit()
}

func (orderStore *Store) CreateWithTx(databaseTransaction *sql.Tx, orderInstance *ordermodel.Order) error {
	orderInstance.ID = uuid.New().String()
	orderInstance.CreatedAt = time.Now()
	orderInstance.Status = "Pending"

	insertionQuery := `INSERT INTO orders (id, user_id, total, status, coupon_code, discount_amount, is_pickup, pickup_time, pickup_location, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, executionError := databaseTransaction.Exec(insertionQuery, orderInstance.ID, orderInstance.UserID, orderInstance.Total, orderInstance.Status, orderInstance.CouponCode, orderInstance.DiscountAmount, orderInstance.IsPickup, orderInstance.PickupTime, orderInstance.PickupLocation, orderInstance.CreatedAt)
	if executionError != nil {
		return executionError
	}

	for _, item := range orderInstance.Items {
		_, itemExecutionError := databaseTransaction.Exec(`INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)`,
			orderInstance.ID, item.ProductID, item.Quantity)
		if itemExecutionError != nil {
			return itemExecutionError
		}
	}
	return nil
}

func (orderStore *Store) GetByID(orderID string) (ordermodel.Order, error) {
	var orderInstance ordermodel.Order
	selectionQuery := `SELECT id, user_id, total, status, coupon_code, discount_amount, is_pickup, pickup_time, pickup_location, items_count, created_at FROM orders WHERE id = $1`
	queryError := orderStore.databaseConnection.QueryRow(selectionQuery, orderID).Scan(
		&orderInstance.ID, &orderInstance.UserID, &orderInstance.Total, &orderInstance.Status,
		&orderInstance.CouponCode, &orderInstance.DiscountAmount, &orderInstance.IsPickup,
		&orderInstance.PickupTime, &orderInstance.PickupLocation, &orderInstance.ItemsCount, &orderInstance.CreatedAt,
	)
	return orderInstance, queryError
}

func (orderStore *Store) GetByUserID(userID int) ([]ordermodel.Order, error) {
	selectionQuery := `SELECT id, user_id, total, status, coupon_code, discount_amount, is_pickup, pickup_time, pickup_location, items_count, created_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC`
	rows, queryError := orderStore.databaseConnection.Query(selectionQuery, userID)
	if queryError != nil {
		return nil, queryError
	}
	defer rows.Close()

	var orderList []ordermodel.Order
	for rows.Next() {
		orderInstance, scanError := orderStore.scanOrder(rows)
		if scanError != nil {
			return nil, scanError
		}

		orderStore.fillOrderItems(&orderInstance)

		orderList = append(orderList, orderInstance)
	}
	return orderList, nil
}

func (orderStore *Store) GetLatestByUserID(userID int) (ordermodel.Order, error) {
	var orderInstance ordermodel.Order
	selectionQuery := `SELECT id, user_id, total, status, coupon_code, discount_amount, is_pickup, pickup_time, pickup_location, items_count, created_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`
	queryError := orderStore.databaseConnection.QueryRow(selectionQuery, userID).Scan(
		&orderInstance.ID, &orderInstance.UserID, &orderInstance.Total, &orderInstance.Status,
		&orderInstance.CouponCode, &orderInstance.DiscountAmount, &orderInstance.IsPickup,
		&orderInstance.PickupTime, &orderInstance.PickupLocation, &orderInstance.ItemsCount, &orderInstance.CreatedAt,
	)
	if queryError == nil {
		orderStore.fillOrderItems(&orderInstance)
	}
	return orderInstance, queryError
}

func (orderStore *Store) GetPickupsByUserID(userID int) ([]ordermodel.Order, error) {
	selectionQuery := `SELECT id, user_id, total, status, coupon_code, discount_amount, is_pickup, pickup_time, pickup_location, items_count, created_at FROM orders WHERE user_id = $1 AND is_pickup = TRUE ORDER BY created_at DESC`
	rows, queryError := orderStore.databaseConnection.Query(selectionQuery, userID)
	if queryError != nil {
		return nil, queryError
	}
	defer rows.Close()

	var orderList []ordermodel.Order
	for rows.Next() {
		orderInstance, scanError := orderStore.scanOrder(rows)
		if scanError != nil {
			return nil, scanError
		}
		orderStore.fillOrderItems(&orderInstance)
		orderList = append(orderList, orderInstance)
	}
	return orderList, nil
}

// Helper methods to avoid repetition
func (orderStore *Store) scanOrder(rows *sql.Rows) (ordermodel.Order, error) {
	var orderInstance ordermodel.Order
	err := rows.Scan(
		&orderInstance.ID, &orderInstance.UserID, &orderInstance.Total, &orderInstance.Status,
		&orderInstance.CouponCode, &orderInstance.DiscountAmount, &orderInstance.IsPickup,
		&orderInstance.PickupTime, &orderInstance.PickupLocation, &orderInstance.ItemsCount, &orderInstance.CreatedAt,
	)
	return orderInstance, err
}

func (orderStore *Store) fillOrderItems(orderInstance *ordermodel.Order) {
	itemRows, err := orderStore.databaseConnection.Query(`SELECT product_id, quantity FROM order_items WHERE order_id = $1`, orderInstance.ID)
	if err == nil {
		defer itemRows.Close()
		for itemRows.Next() {
			var orderItem ordermodel.OrderItem
			itemRows.Scan(&orderItem.ProductID, &orderItem.Quantity)
			orderInstance.Items = append(orderInstance.Items, orderItem)
		}
	}
}

func (orderStore *Store) GetAll() ([]ordermodel.Order, error) {
	selectionQuery := `SELECT id, user_id, total, status, coupon_code, discount_amount, is_pickup, pickup_time, pickup_location, items_count, created_at FROM orders ORDER BY created_at DESC`
	rows, queryError := orderStore.databaseConnection.Query(selectionQuery)
	if queryError != nil {
		return nil, queryError
	}
	defer rows.Close()

	var orderList []ordermodel.Order
	for rows.Next() {
		orderInstance, scanError := orderStore.scanOrder(rows)
		if scanError != nil {
			return nil, scanError
		}
		orderList = append(orderList, orderInstance)
	}
	return orderList, nil
}

func (orderStore *Store) UpdateStatus(orderID string, status string) error {
	updateQuery := "UPDATE orders SET status = $1 WHERE id = $2"
	_, executionError := orderStore.databaseConnection.Exec(updateQuery, status, orderID)
	return executionError
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

func (orderStore *Store) GetDashboardStats() (DashboardStats, error) {
	var dashboardStats DashboardStats

	// 1. Basic Stats
	statsQuery := `
		SELECT 
			COUNT(*), 
			COALESCE(SUM(total), 0),
			COALESCE(AVG(total), 0),
			COUNT(*) FILTER (WHERE status = 'Pending')
		FROM orders
	`
	queryError := orderStore.databaseConnection.QueryRow(statsQuery).Scan(
		&dashboardStats.TotalOrders, &dashboardStats.TotalRevenue,
		&dashboardStats.AverageOrderValue, &dashboardStats.PendingOrders,
	)
	if queryError != nil {
		return dashboardStats, queryError
	}

	// 1.5. Calculate Total Takeout/Pickup orders
	orderStore.databaseConnection.QueryRow(`SELECT COUNT(*) FROM orders WHERE is_pickup = TRUE`).Scan(&dashboardStats.TakeoutOrders)

	// 2. Sales History (last 7 days)
	historyQuery := `
		SELECT TO_CHAR(created_at, 'YYYY-MM-DD') as day, SUM(total), COUNT(*) FILTER (WHERE is_pickup = TRUE)
		FROM orders 
		GROUP BY day 
		ORDER BY day ASC 
		LIMIT 7
	`
	rows, historyQueryError := orderStore.databaseConnection.Query(historyQuery)
	if historyQueryError == nil {
		defer rows.Close()
		for rows.Next() {
			var dailySaleInstance DailySale
			if scanError := rows.Scan(&dailySaleInstance.Date, &dailySaleInstance.Total, &dailySaleInstance.PickupCount); scanError == nil {
				dashboardStats.SalesHistory = append(dashboardStats.SalesHistory, dailySaleInstance)
			}
		}
	}

	return dashboardStats, nil
}
