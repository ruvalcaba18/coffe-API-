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

func (s *Store) GetAll(f productmodel.Filter) ([]productmodel.Product, error) {
	ctx := context.Background()
	
	// Generate dynamic cache key based on filters
	cacheKey := fmt.Sprintf("products:q=%s:c=%s:min=%.2f:max=%.2f", f.Query, f.Category, f.MinPrice, f.MaxPrice)

	// Try cache first
	if val, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil {
		var products []productmodel.Product
		if err := json.Unmarshal([]byte(val), &products); err == nil {
			return products, nil
		}
	}

	// Dynamic Query Construction
	query := "SELECT id, name, description, price, category FROM products WHERE 1=1"
	var args []interface{}
	argCount := 1

	if f.Query != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argCount, argCount+1)
		args = append(args, "%"+f.Query+"%", "%"+f.Query+"%")
		argCount += 2
	}
	if f.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, f.Category)
		argCount++
	}
	if f.MinPrice > 0 {
		query += fmt.Sprintf(" AND price >= $%d", argCount)
		args = append(args, f.MinPrice)
		argCount++
	}
	if f.MaxPrice > 0 {
		query += fmt.Sprintf(" AND price <= $%d", argCount)
		args = append(args, f.MaxPrice)
		argCount++
	}

	rows, err := s.db.Query(query, args...)
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

	// Save to cache for 5 minutes (shorter for filtered queries)
	if data, err := json.Marshal(products); err == nil {
		s.rdb.Set(ctx, cacheKey, data, 5*time.Minute)
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
