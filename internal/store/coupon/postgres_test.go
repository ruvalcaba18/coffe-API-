package coupon

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestStore_IncrementUsage(t *testing.T) {
	databaseMock, sqlMock, error := sqlmock.New()
	if error != nil {
		t.Fatalf("failed to open sqlmock: %s", error)
	}
	defer databaseMock.Close()

	couponStore := NewStore(databaseMock)

	code := "PROMO2024"

	sqlMock.ExpectExec(regexp.QuoteMeta("UPDATE coupons SET used_count = used_count + 1 WHERE code = $1")).
		WithArgs(code).
		WillReturnResult(sqlmock.NewResult(0, 1))

	error = couponStore.IncrementUsage(context.Background(), nil, code)
	assert.NoError(t, error)

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta("UPDATE coupons SET used_count = used_count + 1 WHERE code = $1")).
		WithArgs(code).
		WillReturnResult(sqlmock.NewResult(0, 1))
	sqlMock.ExpectCommit()

	transaction, _ := databaseMock.Begin()
	error = couponStore.IncrementUsage(context.Background(), transaction, code)
	transaction.Commit()
	assert.NoError(t, error)

	if error := sqlMock.ExpectationsWereMet(); error != nil {
		t.Errorf("there were unfulfilled expectations: %s", error)
	}
}
