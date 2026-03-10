package product

import (
	productmodel "coffeebase-api/internal/models/product"
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestStore_GetByID(t *testing.T) {
	// Setup SQL mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	// Setup MiniRedis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to run miniredis: %s", err)
	}
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	s := NewStore(db, redisClient)

	productID := 1
	expectedProduct := productmodel.Product{
		ID:            productID,
		Name:          "Espresso",
		Description:   "Strong coffee",
		Price:         3.50,
		Category:      "Coffee",
		AverageRating: 4.8,
		ReviewCount:   120,
	}

	// 1. First call: Cache miss, should query DB and save to cache
	rows := sqlmock.NewRows([]string{"id", "name", "description", "price", "category", "average_rating", "review_count"}).
		AddRow(expectedProduct.ID, expectedProduct.Name, expectedProduct.Description, expectedProduct.Price, expectedProduct.Category, expectedProduct.AverageRating, expectedProduct.ReviewCount)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, price, category, average_rating, review_count FROM products WHERE id = $1")).
		WithArgs(productID).
		WillReturnRows(rows)

	product, err := s.GetByID(productID)
	assert.NoError(t, err)
	assert.Equal(t, expectedProduct.Name, product.Name)

	// Verify it's in redis
	assert.True(t, mr.Exists("product:1"))

	// 2. Second call: Cache hit, should NOT query DB
	product2, err := s.GetByID(productID)
	assert.NoError(t, err)
	assert.Equal(t, expectedProduct.Name, product2.Name)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to run miniredis: %s", err)
	}
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	s := NewStore(db, redisClient)

	p := &productmodel.Product{
		Name:        "Latte",
		Description: "Milk coffee",
		Price:       4.00,
		Category:    "Coffee",
	}

	// Set a key to verify invalidation
	redisClient.Set(context.Background(), "all_products", "some data", 0)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO products (name, description, price, category) VALUES ($1, $2, $3, $4) RETURNING id")).
		WithArgs(p.Name, p.Description, p.Price, p.Category).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	err = s.Create(p)
	assert.NoError(t, err)
	assert.Equal(t, 2, p.ID)

	// Verify cache invalidation
	assert.False(t, mr.Exists("all_products"))

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
