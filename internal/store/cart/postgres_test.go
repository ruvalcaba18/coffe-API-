package cart

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestStore_GetCart(t *testing.T) {
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

	userID := 1
	
	// 1. Cache miss
	mock.ExpectQuery(regexp.QuoteMeta("SELECT product_id, quantity FROM cart_items WHERE user_id = $1")).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"product_id", "quantity"}).
			AddRow(101, 2).
			AddRow(102, 1))

	cart, err := s.GetCart(userID)
	assert.NoError(t, err)
	assert.Len(t, cart.Items, 2)
	assert.Equal(t, 101, cart.Items[0].ProductID)
	assert.Equal(t, 2, cart.Items[0].Quantity)

	// Verify cached in Redis
	assert.True(t, mr.Exists("cart:1"))

	// 2. Cache hit
	cart2, err := s.GetCart(userID)
	assert.NoError(t, err)
	assert.Len(t, cart2.Items, 2)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestStore_ClearCart(t *testing.T) {
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

	userID := 1

	// Setup Redis key
	redisClient.Set(context.Background(), "cart:1", "data", 0)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM cart_items WHERE user_id = $1")).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tx, _ := db.Begin()
	err = s.ClearCart(tx, userID)
	tx.Commit()

	assert.NoError(t, err)
	assert.False(t, mr.Exists("cart:1"))

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
