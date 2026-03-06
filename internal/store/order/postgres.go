package order

import (
	ordermodel "coffeebase-api/internal/models/order"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(o *ordermodel.Order) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.CreateWithTx(tx, o); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) CreateWithTx(tx *sql.Tx, o *ordermodel.Order) error {
	o.ID = uuid.New().String()
	o.CreatedAt = time.Now()
	o.Status = "Pending"

	query := `INSERT INTO orders (id, user_id, total, status, coupon_code, discount_amount, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := tx.Exec(query, o.ID, o.UserID, o.Total, o.Status, o.CouponCode, o.DiscountAmount, o.CreatedAt)
	if err != nil {
		return err
	}

	for _, item := range o.Items {
		_, err = tx.Exec(`INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)`,
			o.ID, item.ProductID, item.Quantity)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetByID(id string) (ordermodel.Order, error) {
	var o ordermodel.Order
	query := `SELECT id, user_id, total, status, coupon_code, discount_amount, created_at FROM orders WHERE id = $1`
	err := s.db.QueryRow(query, id).Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.CouponCode, &o.DiscountAmount, &o.CreatedAt)
	return o, err
}

func (s *Store) GetByUserID(userID int) ([]ordermodel.Order, error) {
	rows, err := s.db.Query(`SELECT id, user_id, total, status, coupon_code, discount_amount, created_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []ordermodel.Order
	for rows.Next() {
		var o ordermodel.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.CouponCode, &o.DiscountAmount, &o.CreatedAt); err != nil {
			return nil, err
		}
		
		// Fill items (simplified for now, ideally join)
		itemRows, _ := s.db.Query(`SELECT product_id, quantity FROM order_items WHERE order_id = $1`, o.ID)
		for itemRows.Next() {
			var i ordermodel.OrderItem
			itemRows.Scan(&i.ProductID, &i.Quantity)
			o.Items = append(o.Items, i)
		}
		itemRows.Close()

		orders = append(orders, o)
	}
	return orders, nil
}
func (s *Store) GetAll() ([]ordermodel.Order, error) {
	rows, err := s.db.Query(`SELECT id, user_id, total, status, coupon_code, discount_amount, created_at FROM orders ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []ordermodel.Order
	for rows.Next() {
		var o ordermodel.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.CouponCode, &o.DiscountAmount, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (s *Store) UpdateStatus(id string, status string) error {
	_, err := s.db.Exec("UPDATE orders SET status = $1 WHERE id = $2", status, id)
	return err
}

type DailySale struct {
	Date  string  `json:"date"`
	Total float64 `json:"total"`
}

type DashboardStats struct {
	TotalOrders       int         `json:"total_orders"`
	TotalRevenue      float64     `json:"total_revenue"`
	AverageOrderValue float64     `json:"avg_order_value"`
	PendingOrders     int         `json:"pending_orders"`
	SalesHistory      []DailySale `json:"sales_history"`
}

func (s *Store) GetDashboardStats() (DashboardStats, error) {
	var stats DashboardStats

	// 1. Basic Stats
	err := s.db.QueryRow(`
		SELECT 
			COUNT(*), 
			COALESCE(SUM(total), 0),
			COALESCE(AVG(total), 0),
			COUNT(*) FILTER (WHERE status = 'Pending')
		FROM orders
	`).Scan(&stats.TotalOrders, &stats.TotalRevenue, &stats.AverageOrderValue, &stats.PendingOrders)
	if err != nil {
		return stats, err
	}

	// 2. Sales History (last 7 days)
	rows, err := s.db.Query(`
		SELECT TO_CHAR(created_at, 'YYYY-MM-DD') as day, SUM(total) 
		FROM orders 
		GROUP BY day 
		ORDER BY day ASC 
		LIMIT 7
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ds DailySale
			if err := rows.Scan(&ds.Date, &ds.Total); err == nil {
				stats.SalesHistory = append(stats.SalesHistory, ds)
			}
		}
	}

	return stats, nil
}
