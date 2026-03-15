package order

import (
	ordermodel "coffeebase-api/internal/models/order"
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// --- Public ---

func (orderStore *postgresStore) Create(requestContext context.Context, orderInstance *ordermodel.Order) error {
	transaction, error := orderStore.databaseConnection.BeginTx(requestContext, nil)
	if error != nil {
		return error
	}
	defer transaction.Rollback()

	if error := orderStore.CreateWithTx(requestContext, transaction, orderInstance); error != nil {
		return error
	}

	return transaction.Commit()
}

func (orderStore *postgresStore) CreateWithTx(requestContext context.Context, transaction *sql.Tx, order *ordermodel.Order) error {
	order.ID = uuid.New().String()
	order.CreatedAt = time.Now()
	order.Status = "Pending"

	query := `INSERT INTO orders (id, user_id, total, status, coupon_code, discount_amount, is_pickup, pickup_time, pickup_location, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, error := transaction.ExecContext(requestContext, query, order.ID, order.UserID, order.Total, order.Status, order.CouponCode, order.DiscountAmount, order.IsPickup, order.PickupTime, order.PickupLocation, order.CreatedAt)
	if error != nil {
		return error
	}

	for _, item := range order.Items {
		_, error := transaction.ExecContext(requestContext, `INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)`,
			order.ID, item.ProductID, item.Quantity)
		if error != nil {
			return error
		}
	}
	return nil
}

func (orderStore *postgresStore) GetByID(requestContext context.Context, id string) (ordermodel.Order, error) {
	var order ordermodel.Order
	query := `SELECT id, user_id, total, status, coupon_code, discount_amount, is_pickup, pickup_time, pickup_location, items_count, created_at FROM orders WHERE id = $1`
	error := orderStore.databaseConnection.QueryRowContext(requestContext, query, id).Scan(
		&order.ID, &order.UserID, &order.Total, &order.Status,
		&order.CouponCode, &order.DiscountAmount, &order.IsPickup,
		&order.PickupTime, &order.PickupLocation, &order.ItemsCount, &order.CreatedAt,
	)
	return order, error
}

func (orderStore *postgresStore) GetByUserID(requestContext context.Context, userID int) ([]ordermodel.Order, error) {
	query := `SELECT id, user_id, total, status, coupon_code, discount_amount, is_pickup, pickup_time, pickup_location, items_count, created_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC`
	return orderStore.processOrderRows(requestContext, query, true, userID)
}

func (orderStore *postgresStore) GetLatestByUserID(requestContext context.Context, userID int) (ordermodel.Order, error) {
	var order ordermodel.Order
	query := `SELECT id, user_id, total, status, coupon_code, discount_amount, is_pickup, pickup_time, pickup_location, items_count, created_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`
	error := orderStore.databaseConnection.QueryRowContext(requestContext, query, userID).Scan(
		&order.ID, &order.UserID, &order.Total, &order.Status,
		&order.CouponCode, &order.DiscountAmount, &order.IsPickup,
		&order.PickupTime, &order.PickupLocation, &order.ItemsCount, &order.CreatedAt,
	)
	if error == nil {
		orderStore.fillOrderItems(requestContext, &order)
	}
	return order, error
}

func (orderStore *postgresStore) GetPickupsByUserID(requestContext context.Context, userID int) ([]ordermodel.Order, error) {
	query := `SELECT id, user_id, total, status, coupon_code, discount_amount, is_pickup, pickup_time, pickup_location, items_count, created_at FROM orders WHERE user_id = $1 AND is_pickup = TRUE ORDER BY created_at DESC`
	return orderStore.processOrderRows(requestContext, query, true, userID)
}

func (orderStore *postgresStore) GetAll(requestContext context.Context) ([]ordermodel.Order, error) {
	query := `SELECT id, user_id, total, status, coupon_code, discount_amount, is_pickup, pickup_time, pickup_location, items_count, created_at FROM orders ORDER BY created_at DESC`
	return orderStore.processOrderRows(requestContext, query, false)
}

func (orderStore *postgresStore) UpdateStatus(requestContext context.Context, id string, status string) error {
	query := "UPDATE orders SET status = $1 WHERE id = $2"
	_, error := orderStore.databaseConnection.ExecContext(requestContext, query, status, id)
	return error
}

func (orderStore *postgresStore) GetDashboardStats(requestContext context.Context) (DashboardStats, error) {
	var stats DashboardStats

	query := `
		SELECT 
			COUNT(*), 
			COALESCE(SUM(total), 0),
			COALESCE(AVG(total), 0),
			COUNT(*) FILTER (WHERE status = 'Pending')
		FROM orders
	`
	error := orderStore.databaseConnection.QueryRowContext(requestContext, query).Scan(
		&stats.TotalOrders, &stats.TotalRevenue,
		&stats.AverageOrderValue, &stats.PendingOrders,
	)
	if error != nil {
		return stats, error
	}

	orderStore.databaseConnection.QueryRowContext(requestContext, `SELECT COUNT(*) FROM orders WHERE is_pickup = TRUE`).Scan(&stats.TakeoutOrders)

	historyQuery := `
		SELECT TO_CHAR(created_at, 'YYYY-MM-DD') as day, SUM(total), COUNT(*) FILTER (WHERE is_pickup = TRUE)
		FROM orders 
		GROUP BY day 
		ORDER BY day ASC 
		LIMIT 7
	`
	rows, error := orderStore.databaseConnection.QueryContext(requestContext, historyQuery)
	if error == nil {
		defer rows.Close()
		for rows.Next() {
			var sale DailySale
			if error := rows.Scan(&sale.Date, &sale.Total, &sale.PickupCount); error == nil {
				stats.SalesHistory = append(stats.SalesHistory, sale)
			}
		}
	}

	return stats, nil
}

// --- Private ---

func (orderStore *postgresStore) processOrderRows(requestContext context.Context, query string, fillItems bool, args ...interface{}) ([]ordermodel.Order, error) {
	rows, error := orderStore.databaseConnection.QueryContext(requestContext, query, args...)
	if error != nil {
		return nil, error
	}
	defer rows.Close()

	var orders []ordermodel.Order
	for rows.Next() {
		orderInstance, error := orderStore.scanOrder(rows)
		if error != nil {
			return nil, error
		}
		if fillItems {
			orderStore.fillOrderItems(requestContext, &orderInstance)
		}
		orders = append(orders, orderInstance)
	}
	return orders, nil
}

func (orderStore *postgresStore) scanOrder(rows *sql.Rows) (ordermodel.Order, error) {
	var order ordermodel.Order
	error := rows.Scan(
		&order.ID, &order.UserID, &order.Total, &order.Status,
		&order.CouponCode, &order.DiscountAmount, &order.IsPickup,
		&order.PickupTime, &order.PickupLocation, &order.ItemsCount, &order.CreatedAt,
	)
	return order, error
}

func (orderStore *postgresStore) fillOrderItems(requestContext context.Context, order *ordermodel.Order) {
	query := `SELECT product_id, quantity FROM order_items WHERE order_id = $1`
	rows, error := orderStore.databaseConnection.QueryContext(requestContext, query, order.ID)
	if error == nil {
		defer rows.Close()
		for rows.Next() {
			var item ordermodel.OrderItem
			rows.Scan(&item.ProductID, &item.Quantity)
			order.Items = append(order.Items, item)
		}
	}
}
