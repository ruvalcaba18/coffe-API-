package product

import (
	productmodel "coffeebase-api/internal/models/product"
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

func (s *Store) GetAll() ([]productmodel.Product, error) {
	ctx := context.Background()
	cacheKey := "all_products"

	// Try cache first
	if val, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil {
		var products []productmodel.Product
		if err := json.Unmarshal([]byte(val), &products); err == nil {
			return products, nil
		}
	}

	// Cache miss: Query DB
	rows, err := s.db.Query("SELECT id, name, description, price, category FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []productmodel.Product
	for rows.Next() {
		var p productmodel.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	// Save to cache for 10 minutes
	if data, err := json.Marshal(products); err == nil {
		s.rdb.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return products, nil
}

func (s *Store) GetByID(id int) (productmodel.Product, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("product:%d", id)

	// Try cache
	if val, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil {
		var p productmodel.Product
		if err := json.Unmarshal([]byte(val), &p); err == nil {
			return p, nil
		}
	}

	// Cache miss
	var p productmodel.Product
	err := s.db.QueryRow("SELECT id, name, description, price, category FROM products WHERE id = $1", id).
		Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category)
	
	if err == nil {
		// Save to cache
		if data, err := json.Marshal(p); err == nil {
			s.rdb.Set(ctx, cacheKey, data, 10*time.Minute)
		}
	}

	return p, err
}

func (s *Store) Create(p *productmodel.Product) error {
	query := `INSERT INTO products (name, description, price, category) VALUES ($1, $2, $3, $4) RETURNING id`
	err := s.db.QueryRow(query, p.Name, p.Description, p.Price, p.Category).Scan(&p.ID)
	if err == nil {
		s.rdb.Del(context.Background(), "all_products")
	}
	return err
}

func (s *Store) Update(p *productmodel.Product) error {
	query := `UPDATE products SET name = $1, description = $2, price = $3, category = $4 WHERE id = $5`
	_, err := s.db.Exec(query, p.Name, p.Description, p.Price, p.Category, p.ID)
	if err == nil {
		ctx := context.Background()
		s.rdb.Del(ctx, "all_products")
		s.rdb.Del(ctx, fmt.Sprintf("product:%d", p.ID))
	}
	return err
}

func (s *Store) Delete(id int) error {
	_, err := s.db.Exec("DELETE FROM products WHERE id = $1", id)
	if err == nil {
		ctx := context.Background()
		s.rdb.Del(ctx, "all_products")
		s.rdb.Del(ctx, fmt.Sprintf("product:%d", id))
	}
	return err
}
