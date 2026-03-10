package coupon

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestStore_IncrementUsage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	s := NewStore(db)

	code := "PROMO2024"

	// Case 1: Without transaction
	mock.ExpectExec(regexp.QuoteMeta("UPDATE coupons SET used_count = used_count + 1 WHERE code = $1")).
		WithArgs(code).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = s.IncrementUsage(nil, code)
	assert.NoError(t, err)

	// Case 2: With transaction
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE coupons SET used_count = used_count + 1 WHERE code = $1")).
		WithArgs(code).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tx, _ := db.Begin()
	err = s.IncrementUsage(tx, code)
	tx.Commit()
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
