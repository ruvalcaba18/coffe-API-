package cart

import (
	"coffeebase-api/internal/cache"
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestStore_GetCart(t *testing.T) {
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

	cartStore := NewStore(databaseMock, cache.NewRedisCache(redisClient))

	userID := 1
	
	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT product_id, quantity FROM cart_items WHERE user_id = $1")).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"product_id", "quantity"}).
			AddRow(101, 2).
			AddRow(102, 1))

	cartInstance, error := cartStore.GetCart(context.Background(), userID)
	assert.NoError(t, error)
	assert.Len(t, cartInstance.Items, 2)
	assert.Equal(t, 101, cartInstance.Items[0].ProductID)
	assert.Equal(t, 2, cartInstance.Items[0].Quantity)

	assert.True(t, miniRedis.Exists("cart:1"))

	cartInstance2, error := cartStore.GetCart(context.Background(), userID)
	assert.NoError(t, error)
	assert.Len(t, cartInstance2.Items, 2)

	if error := sqlMock.ExpectationsWereMet(); error != nil {
		t.Errorf("there were unfulfilled expectations: %s", error)
	}
}

func TestStore_ClearCart(t *testing.T) {
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

	cartStore := NewStore(databaseMock, cache.NewRedisCache(redisClient))

	userID := 1

	redisClient.Set(context.Background(), "cart:1", "data", 0)

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta("DELETE FROM cart_items WHERE user_id = $1")).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	sqlMock.ExpectCommit()

	transaction, _ := databaseMock.Begin()
	error = cartStore.ClearCart(context.Background(), transaction, userID)
	transaction.Commit()

	assert.NoError(t, error)
	assert.False(t, miniRedis.Exists("cart:1"))

	if error := sqlMock.ExpectationsWereMet(); error != nil {
		t.Errorf("there were unfulfilled expectations: %s", error)
	}
}
