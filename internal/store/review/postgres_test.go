package review

import (
	reviewmodel "coffeebase-api/internal/models/review"
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestStore_Create(t *testing.T) {
	databaseMock, sqlMock, mockError := sqlmock.New()
	if mockError != nil {
		t.Fatalf("failed to open sqlmock: %s", mockError)
	}
	defer databaseMock.Close()

	reviewStore := NewStore(databaseMock)

	now := time.Now()
	reviewInstance := &reviewmodel.Review{
		ProductID: 1,
		UserID:    5,
		Rating:    4,
		Comment:   "Great coffee!",
	}

	sqlMock.ExpectQuery(regexp.QuoteMeta("INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) RETURNING id, created_at")).
		WithArgs(reviewInstance.ProductID, reviewInstance.UserID, reviewInstance.Rating, reviewInstance.Comment).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now))

	insertError := reviewStore.Create(context.Background(), reviewInstance)
	assert.NoError(t, insertError)
	assert.Equal(t, 1, reviewInstance.ID)
	assert.Equal(t, now, reviewInstance.CreatedAt)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestStore_GetByProductID(t *testing.T) {
	databaseMock, sqlMock, mockError := sqlmock.New()
	if mockError != nil {
		t.Fatalf("failed to open sqlmock: %s", mockError)
	}
	defer databaseMock.Close()

	reviewStore := NewStore(databaseMock)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "product_id", "user_id", "rating", "comment", "created_at"}).
		AddRow(1, 1, 5, 5, "Amazing!", now).
		AddRow(2, 1, 6, 4, "Pretty good", now)

	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT id, product_id, user_id, rating, comment, created_at FROM reviews WHERE product_id = $1")).
		WithArgs(1).
		WillReturnRows(rows)

	reviews, queryError := reviewStore.GetByProductID(context.Background(), 1)
	assert.NoError(t, queryError)
	assert.Len(t, reviews, 2)
	assert.Equal(t, 5, reviews[0].Rating)
	assert.Equal(t, "Amazing!", reviews[0].Comment)
	assert.Equal(t, 4, reviews[1].Rating)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestStore_GetByProductID_Empty(t *testing.T) {
	databaseMock, sqlMock, mockError := sqlmock.New()
	if mockError != nil {
		t.Fatalf("failed to open sqlmock: %s", mockError)
	}
	defer databaseMock.Close()

	reviewStore := NewStore(databaseMock)

	rows := sqlmock.NewRows([]string{"id", "product_id", "user_id", "rating", "comment", "created_at"})

	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT id, product_id, user_id, rating, comment, created_at FROM reviews WHERE product_id = $1")).
		WithArgs(999).
		WillReturnRows(rows)

	reviews, queryError := reviewStore.GetByProductID(context.Background(), 999)
	assert.NoError(t, queryError)
	assert.Empty(t, reviews)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
