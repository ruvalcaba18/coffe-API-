package cart

import (
	cartmodel "coffeebase-api/internal/models/cart"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	db  *sql.DB
	rdb *redis.Client
}

func NewStore(db *sql.DB, rdb *redis.Client) *Store {
	return &Store{
		db:  db,
		rdb: rdb,
	}
}

func (s *Store) UpdateItem(userID, productID, quantity int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Ensure cart exists
	_, err = tx.Exec(`INSERT INTO carts (user_id, updated_at) VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE SET updated_at = $2`, userID, time.Now())
	if err != nil {
		return err
	}

	if quantity <= 0 {
		_, err = tx.Exec(`DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2`, userID, productID)
	} else {
		_, err = tx.Exec(`INSERT INTO cart_items (user_id, product_id, quantity) VALUES ($1, $2, $3) 
						 ON CONFLICT (user_id, product_id) DO UPDATE SET quantity = $3`, userID, productID, quantity)
	}
	
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// Invalidate Redis cache
	ctx := context.Background()
	cacheKey := fmt.Sprintf("cart:%d", userID)
	s.rdb.Del(ctx, cacheKey)

	return nil
}

func (s *Store) GetCart(userID int) (*cartmodel.Cart, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("cart:%d", userID)

	// Try Redis
	if val, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil {
		var c cartmodel.Cart
		if err := json.Unmarshal([]byte(val), &c); err == nil {
			return &c, nil
		}
	}

	// Cache miss: Postgres
	rows, err := s.db.Query(`SELECT product_id, quantity FROM cart_items WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	c := &cartmodel.Cart{UserID: userID, Items: []cartmodel.Item{}}
	for rows.Next() {
		var item cartmodel.Item
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			return nil, err
		}
		c.Items = append(c.Items, item)
	}

	// Save to Redis (expire in 1 hour if inactive)
	if data, err := json.Marshal(c); err == nil {
		s.rdb.Set(ctx, cacheKey, data, 1*time.Hour)
	}

	return c, nil
}

func (s *Store) ClearCart(tx *sql.Tx, userID int) error {
	_, err := tx.Exec(`DELETE FROM cart_items WHERE user_id = $1`, userID)
	if err == nil {
		// Also clear Redis
		ctx := context.Background()
		cacheKey := fmt.Sprintf("cart:%d", userID)
		s.rdb.Del(ctx, cacheKey)
	}
	return err
}
