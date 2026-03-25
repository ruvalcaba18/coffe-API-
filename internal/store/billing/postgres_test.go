package billing

import (
	billingmodel "coffeebase-api/internal/models/billing"
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestStore_GetPaymentMethodsByUserID(t *testing.T) {
	databaseMock, sqlMock, mockError := sqlmock.New()
	if mockError != nil {
		t.Fatalf("failed to open sqlmock: %s", mockError)
	}
	defer databaseMock.Close()

	billingStore := NewStore(databaseMock)

	userID := 1
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "last4", "expiry", "brand", "holder", "is_default", "created_at"}).
		AddRow(1, userID, "4242", "12/26", "Visa", "John Doe", true, now).
		AddRow(2, userID, "5555", "09/25", "Mastercard", "John Doe", false, now)

	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, last4, expiry, brand, holder, is_default, created_at FROM payment_methods WHERE user_id = $1")).
		WithArgs(userID).
		WillReturnRows(rows)

	paymentMethods, queryError := billingStore.GetPaymentMethodsByUserID(context.Background(), userID)
	assert.NoError(t, queryError)
	assert.Len(t, paymentMethods, 2)
	assert.Equal(t, "4242", paymentMethods[0].Last4)
	assert.True(t, paymentMethods[0].IsDefault)
	assert.Equal(t, "5555", paymentMethods[1].Last4)
	assert.False(t, paymentMethods[1].IsDefault)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestStore_GetPaymentMethodsByUserID_Empty(t *testing.T) {
	databaseMock, sqlMock, mockError := sqlmock.New()
	if mockError != nil {
		t.Fatalf("failed to open sqlmock: %s", mockError)
	}
	defer databaseMock.Close()

	billingStore := NewStore(databaseMock)

	rows := sqlmock.NewRows([]string{"id", "user_id", "last4", "expiry", "brand", "holder", "is_default", "created_at"})

	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, last4, expiry, brand, holder, is_default, created_at FROM payment_methods WHERE user_id = $1")).
		WithArgs(99).
		WillReturnRows(rows)

	paymentMethods, queryError := billingStore.GetPaymentMethodsByUserID(context.Background(), 99)
	assert.NoError(t, queryError)
	assert.Empty(t, paymentMethods)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestStore_AddPaymentMethod(t *testing.T) {
	databaseMock, sqlMock, mockError := sqlmock.New()
	if mockError != nil {
		t.Fatalf("failed to open sqlmock: %s", mockError)
	}
	defer databaseMock.Close()

	billingStore := NewStore(databaseMock)

	now := time.Now()
	paymentMethod := &billingmodel.PaymentMethod{
		UserID: 1,
		Last4:  "4242",
		Expiry: "12/26",
		Brand:  "Visa",
		Holder: "John Doe",
	}

	sqlMock.ExpectQuery(regexp.QuoteMeta("INSERT INTO payment_methods (user_id, last4, expiry, brand, holder, is_default) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at")).
		WithArgs(paymentMethod.UserID, paymentMethod.Last4, paymentMethod.Expiry, paymentMethod.Brand, paymentMethod.Holder, paymentMethod.IsDefault).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now))

	insertError := billingStore.AddPaymentMethod(context.Background(), paymentMethod)
	assert.NoError(t, insertError)
	assert.Equal(t, 1, paymentMethod.ID)
	assert.Equal(t, now, paymentMethod.CreatedAt)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestStore_ExistsPaymentMethod_Found(t *testing.T) {
	databaseMock, sqlMock, mockError := sqlmock.New()
	if mockError != nil {
		t.Fatalf("failed to open sqlmock: %s", mockError)
	}
	defer databaseMock.Close()

	billingStore := NewStore(databaseMock)

	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM payment_methods WHERE user_id = $1 AND last4 = $2 AND brand = $3)")).
		WithArgs(1, "4242", "Visa").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, queryError := billingStore.ExistsPaymentMethod(context.Background(), 1, "4242", "Visa")
	assert.NoError(t, queryError)
	assert.True(t, exists)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestStore_ExistsPaymentMethod_NotFound(t *testing.T) {
	databaseMock, sqlMock, mockError := sqlmock.New()
	if mockError != nil {
		t.Fatalf("failed to open sqlmock: %s", mockError)
	}
	defer databaseMock.Close()

	billingStore := NewStore(databaseMock)

	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM payment_methods WHERE user_id = $1 AND last4 = $2 AND brand = $3)")).
		WithArgs(1, "9999", "Visa").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, queryError := billingStore.ExistsPaymentMethod(context.Background(), 1, "9999", "Visa")
	assert.NoError(t, queryError)
	assert.False(t, exists)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
