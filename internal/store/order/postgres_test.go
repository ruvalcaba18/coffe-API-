package order

import (
	ordermodel "coffeebase-api/internal/models/order"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	s := NewStore(db)

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

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO orders (id, user_id, total, status, coupon_code, discount_amount, is_pickup, pickup_time, pickup_location, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)")).
		WithArgs(sqlmock.AnyArg(), orderInstance.UserID, orderInstance.Total, "Pending", "", 0.0, orderInstance.IsPickup, nil, orderInstance.PickupLocation, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)")).
		WithArgs(sqlmock.AnyArg(), 101, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)")).
		WithArgs(sqlmock.AnyArg(), 102, 2).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = s.Create(orderInstance)
	assert.NoError(t, err)
	assert.NotEmpty(t, orderInstance.ID)
	assert.Equal(t, "Pending", orderInstance.Status)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestStore_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	s := NewStore(db)

	orderID := "some-uuid"
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "total", "status", "coupon_code", "discount_amount", "is_pickup", "pickup_time", "pickup_location", "items_count", "created_at"}).
		AddRow(orderID, 1, 10.50, "Completed", "PROMO", 1.0, true, now, "Store 1", 2, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, total, status, coupon_code, discount_amount, is_pickup, pickup_time, pickup_location, items_count, created_at FROM orders WHERE id = $1")).
		WithArgs(orderID).
		WillReturnRows(rows)

	order, err := s.GetByID(orderID)
	assert.NoError(t, err)
	assert.Equal(t, orderID, order.ID)
	assert.Equal(t, "Completed", order.Status)
	assert.Equal(t, "PROMO", order.CouponCode)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
