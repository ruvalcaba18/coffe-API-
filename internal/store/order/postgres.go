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

	query := `INSERT INTO orders (id, user_id, total, status, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := tx.Exec(query, o.ID, o.UserID, o.Total, o.Status, o.CreatedAt)
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

func (s *Store) GetByUserID(userID int) ([]ordermodel.Order, error) {
	rows, err := s.db.Query(`SELECT id, user_id, total, status, created_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []ordermodel.Order
	for rows.Next() {
		var o ordermodel.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.CreatedAt); err != nil {
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
