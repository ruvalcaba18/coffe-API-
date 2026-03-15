package product

import (
	"coffeebase-api/internal/cache"
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
	databaseMock, sqlMock, error := sqlmock.New()
	if error != nil {
		t.Fatalf("failed to open sqlmock: %s", error)
	}
	defer databaseMock.Close()

	miniRedis, error := miniredis.Run()
	if error != nil {
		t.Fatalf("failed to run miniredis: %s", error)
	}
	defer miniRedis.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: miniRedis.Addr(),
	})

	productStore := NewStore(databaseMock, cache.NewRedisCache(redisClient))

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

	rows := sqlmock.NewRows([]string{"id", "name", "description", "price", "category", "average_rating", "review_count"}).
		AddRow(expectedProduct.ID, expectedProduct.Name, expectedProduct.Description, expectedProduct.Price, expectedProduct.Category, expectedProduct.AverageRating, expectedProduct.ReviewCount)

	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, price, category, average_rating, review_count FROM products WHERE id = $1")).
		WithArgs(productID).
		WillReturnRows(rows)

	productInstance, error := productStore.GetByID(context.Background(), productID)
	assert.NoError(t, error)
	assert.Equal(t, expectedProduct.Name, productInstance.Name)
	assert.True(t, miniRedis.Exists("product:1"))

	productInstance2, error := productStore.GetByID(context.Background(), productID)
	assert.NoError(t, error)
	assert.Equal(t, expectedProduct.Name, productInstance2.Name)

	if error := sqlMock.ExpectationsWereMet(); error != nil {
		t.Errorf("there were unfulfilled expectations: %s", error)
	}
}

func TestStore_Create(t *testing.T) {
	databaseMock, sqlMock, error := sqlmock.New()
	if error != nil {
		t.Fatalf("failed to open sqlmock: %s", error)
	}
	defer databaseMock.Close()

	miniRedis, error := miniredis.Run()
	if error != nil {
		t.Fatalf("failed to run miniredis: %s", error)
	}
	defer miniRedis.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: miniRedis.Addr(),
	})

	productStore := NewStore(databaseMock, cache.NewRedisCache(redisClient))

	productInstance := &productmodel.Product{
		Name:        "Latte",
		Description: "Milk coffee",
		Price:       4.00,
		Category:    "Coffee",
	}

	redisClient.Set(context.Background(), "all_products", "some data", 0)

	sqlMock.ExpectQuery(regexp.QuoteMeta("INSERT INTO products (name, description, price, category) VALUES ($1, $2, $3, $4) RETURNING id")).
		WithArgs(productInstance.Name, productInstance.Description, productInstance.Price, productInstance.Category).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	error = productStore.Create(context.Background(), productInstance)
	assert.NoError(t, error)
	assert.Equal(t, 2, productInstance.ID)

	assert.False(t, miniRedis.Exists("all_products"))

	if error := sqlMock.ExpectationsWereMet(); error != nil {
		t.Errorf("there were unfulfilled expectations: %s", error)
	}
}
