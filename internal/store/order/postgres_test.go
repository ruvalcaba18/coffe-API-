package order

import (
	ordermodel "coffeebase-api/internal/models/order"
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestStore_Create(t *testing.T) {
	databaseMock, sqlMock, error := sqlmock.New()
	if error != nil {
		t.Fatalf("failed to open sqlmock: %s", error)
	}
	defer databaseMock.Close()

	orderStore := NewStore(databaseMock)

	orderInstance := &ordermodel.Order{
		UserID:         1,
		Total:          10.50,
		IsPickup:       true,
		PickupLocation: "Main Street",
		Items: []ordermodel.OrderItem{
			{ProductID: 101, Quantity: 1},
			{ProductID: 102, Quantity: 2},
		},
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta("INSERT INTO orders (id, user_id, total, status, coupon_code, discount_amount, is_pickup, pickup_time, pickup_location, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)")).
		WithArgs(sqlmock.AnyArg(), orderInstance.UserID, orderInstance.Total, "Pending", "", 0.0, orderInstance.IsPickup, nil, orderInstance.PickupLocation, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	sqlMock.ExpectExec(regexp.QuoteMeta("INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)")).
		WithArgs(sqlmock.AnyArg(), 101, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	
	sqlMock.ExpectExec(regexp.QuoteMeta("INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)")).
		WithArgs(sqlmock.AnyArg(), 102, 2).
		WillReturnResult(sqlmock.NewResult(1, 1))

	sqlMock.ExpectCommit()

	error = orderStore.Create(context.Background(), orderInstance)
	assert.NoError(t, error)
	assert.NotEmpty(t, orderInstance.ID)
	assert.Equal(t, "Pending", orderInstance.Status)

	if error := sqlMock.ExpectationsWereMet(); error != nil {
		t.Errorf("there were unfulfilled expectations: %s", error)
	}
}

func TestStore_GetByID(t *testing.T) {
	databaseMock, sqlMock, error := sqlmock.New()
	if error != nil {
		t.Fatalf("failed to open sqlmock: %s", error)
	}
	defer databaseMock.Close()

	orderStore := NewStore(databaseMock)

	orderID := "some-uuid"
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "total", "status", "coupon_code", "discount_amount", "is_pickup", "pickup_time", "pickup_location", "items_count", "created_at"}).
		AddRow(orderID, 1, 10.50, "Completed", "PROMO", 1.0, true, now, "Store 1", 2, now)

	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, total, status, coupon_code, discount_amount, is_pickup, pickup_time, pickup_location, items_count, created_at FROM orders WHERE id = $1")).
		WithArgs(orderID).
		WillReturnRows(rows)

	orderInstance, error := orderStore.GetByID(context.Background(), orderID)
	assert.NoError(t, error)
	assert.Equal(t, orderID, orderInstance.ID)
	assert.Equal(t, "Completed", orderInstance.Status)
	assert.Equal(t, "PROMO", orderInstance.CouponCode)

	if error := sqlMock.ExpectationsWereMet(); error != nil {
		t.Errorf("there were unfulfilled expectations: %s", error)
	}
}
