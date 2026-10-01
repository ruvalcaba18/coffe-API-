package order

import (
	ordermodel "coffeebase-api/internal/models/order"
	"context"
	"database/sql"
	"encoding/json"
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
	query := orderStore.baseOrderQuery() + ` WHERE o.id = $1 GROUP BY o.id`
	row := orderStore.databaseConnection.QueryRowContext(requestContext, query, id)
	return orderStore.scanOrderWithItems(row)
}

func (orderStore *postgresStore) GetByUserID(requestContext context.Context, userID int) ([]ordermodel.Order, error) {
	query := orderStore.baseOrderQuery() + ` WHERE o.user_id = $1 GROUP BY o.id ORDER BY o.created_at DESC`
	return orderStore.queryOrders(requestContext, query, userID)
}

func (orderStore *postgresStore) GetLatestByUserID(requestContext context.Context, userID int) (ordermodel.Order, error) {
	query := orderStore.baseOrderQuery() + ` WHERE o.user_id = $1 GROUP BY o.id ORDER BY o.created_at DESC LIMIT 1`
	row := orderStore.databaseConnection.QueryRowContext(requestContext, query, userID)
	return orderStore.scanOrderWithItems(row)
}

func (orderStore *postgresStore) GetPickupsByUserID(requestContext context.Context, userID int) ([]ordermodel.Order, error) {
	query := orderStore.baseOrderQuery() + ` WHERE o.user_id = $1 AND o.is_pickup = TRUE GROUP BY o.id ORDER BY o.created_at DESC`
	return orderStore.queryOrders(requestContext, query, userID)
}

func (orderStore *postgresStore) GetAll(requestContext context.Context) ([]ordermodel.Order, error) {
	query := orderStore.baseOrderQuery() + ` GROUP BY o.id ORDER BY o.created_at DESC`
	return orderStore.queryOrders(requestContext, query)
}

func (orderStore *postgresStore) UpdateStatus(requestContext context.Context, id string, status string) error {
	query := "UPDATE orders SET status = $1 WHERE id = $2"
	_, error := orderStore.databaseConnection.ExecContext(requestContext, query, status, id)
	return error
}

// GetDashboardStats obtiene todas las métricas del dashboard en una sola query SQL.
// La segunda query de historial usa TO_CHAR + GROUP BY nativamente en PostgreSQL.
func (orderStore *postgresStore) GetDashboardStats(requestContext context.Context) (DashboardStats, error) {
	var stats DashboardStats

	// Una sola query para todas las métricas de órdenes — sin queries separadas en Go
	query := `
		SELECT
			COUNT(*)                                          AS total_orders,
			COALESCE(SUM(total), 0)                          AS total_revenue,
			COALESCE(AVG(total), 0)                          AS avg_order_value,
			COUNT(*) FILTER (WHERE status = 'Pending')       AS pending_orders,
			COUNT(*) FILTER (WHERE is_pickup = TRUE)         AS takeout_orders
		FROM orders`

	error := orderStore.databaseConnection.QueryRowContext(requestContext, query).Scan(
		&stats.TotalOrders, &stats.TotalRevenue,
		&stats.AverageOrderValue, &stats.PendingOrders, &stats.TakeoutOrders,
	)
	if error != nil {
		return stats, error
	}

	// Historial de ventas de los últimos 7 días — calculado en PostgreSQL
	historyQuery := `
		SELECT
			TO_CHAR(DATE_TRUNC('day', created_at), 'YYYY-MM-DD') AS day,
			SUM(total)                                            AS total,
			COUNT(*) FILTER (WHERE is_pickup = TRUE)             AS pickup_count
		FROM orders
		WHERE created_at >= NOW() - INTERVAL '7 days'
		GROUP BY DATE_TRUNC('day', created_at)
		ORDER BY day ASC`

	rows, error := orderStore.databaseConnection.QueryContext(requestContext, historyQuery)
	if error != nil {
		return stats, nil // historial no crítico
	}
	defer rows.Close()

	for rows.Next() {
		var sale DailySale
		if err := rows.Scan(&sale.Date, &sale.Total, &sale.PickupCount); err == nil {
			stats.SalesHistory = append(stats.SalesHistory, sale)
		}
	}

	return stats, nil
}

// --- Private ---

// baseOrderQuery retorna el SELECT base con JOIN a order_items y json_agg.
// Elimina el N+1: los items de cada orden se traen en la misma query con json_agg.
func (orderStore *postgresStore) baseOrderQuery() string {
	return `
		SELECT
			o.id, o.user_id, o.total, o.status,
			o.coupon_code, o.discount_amount, o.is_pickup,
			o.pickup_time, o.pickup_location, o.items_count, o.created_at,
			COALESCE(
				json_agg(
					json_build_object(
						'product_id',   oi.product_id,
						'quantity',     oi.quantity,
						'product_name', p.name
					) ORDER BY oi.product_id
				) FILTER (WHERE oi.product_id IS NOT NULL),
				'[]'
			) AS items
		FROM orders o
		LEFT JOIN order_items oi ON oi.order_id = o.id
		LEFT JOIN products    p  ON p.id = oi.product_id`
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

// scanOrderWithItems escanea una fila con el JSON de items ya incluido.
func (orderStore *postgresStore) scanOrderWithItems(scanner rowScanner) (ordermodel.Order, error) {
	var order ordermodel.Order
	var itemsJSON []byte

	error := scanner.Scan(
		&order.ID, &order.UserID, &order.Total, &order.Status,
		&order.CouponCode, &order.DiscountAmount, &order.IsPickup,
		&order.PickupTime, &order.PickupLocation, &order.ItemsCount, &order.CreatedAt,
		&itemsJSON,
	)
	if error != nil {
		return order, error
	}

	if len(itemsJSON) > 0 {
		json.Unmarshal(itemsJSON, &order.Items)
	}
	return order, nil
}

// queryOrders ejecuta una query de lista y escanea todas las filas.
func (orderStore *postgresStore) queryOrders(requestContext context.Context, query string, args ...interface{}) ([]ordermodel.Order, error) {
	rows, error := orderStore.databaseConnection.QueryContext(requestContext, query, args...)
	if error != nil {
		return nil, error
	}
	defer rows.Close()

	orders := make([]ordermodel.Order, 0)
	for rows.Next() {
		var order ordermodel.Order
		var itemsJSON []byte

		if err := rows.Scan(
			&order.ID, &order.UserID, &order.Total, &order.Status,
			&order.CouponCode, &order.DiscountAmount, &order.IsPickup,
			&order.PickupTime, &order.PickupLocation, &order.ItemsCount, &order.CreatedAt,
			&itemsJSON,
		); err != nil {
			return nil, err
		}

		if len(itemsJSON) > 0 {
			json.Unmarshal(itemsJSON, &order.Items)
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}
