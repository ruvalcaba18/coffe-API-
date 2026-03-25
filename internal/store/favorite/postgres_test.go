package favorite

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestStore_Add(t *testing.T) {
	databaseMock, sqlMock, mockError := sqlmock.New()
	if mockError != nil {
		t.Fatalf("failed to open sqlmock: %s", mockError)
	}
	defer databaseMock.Close()

	favoriteStore := NewStore(databaseMock)

	sqlMock.ExpectExec(regexp.QuoteMeta("INSERT INTO favorites (user_id, product_id) VALUES ($1, $2) ON CONFLICT DO NOTHING")).
		WithArgs(1, 10).
		WillReturnResult(sqlmock.NewResult(0, 1))

	insertError := favoriteStore.Add(context.Background(), 1, 10)
	assert.NoError(t, insertError)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestStore_Remove(t *testing.T) {
	databaseMock, sqlMock, mockError := sqlmock.New()
	if mockError != nil {
		t.Fatalf("failed to open sqlmock: %s", mockError)
	}
	defer databaseMock.Close()

	favoriteStore := NewStore(databaseMock)

	sqlMock.ExpectExec(regexp.QuoteMeta("DELETE FROM favorites WHERE user_id = $1 AND product_id = $2")).
		WithArgs(1, 10).
		WillReturnResult(sqlmock.NewResult(0, 1))

	deleteError := favoriteStore.Remove(context.Background(), 1, 10)
	assert.NoError(t, deleteError)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestStore_GetUserFavorites(t *testing.T) {
	databaseMock, sqlMock, mockError := sqlmock.New()
	if mockError != nil {
		t.Fatalf("failed to open sqlmock: %s", mockError)
	}
	defer databaseMock.Close()

	favoriteStore := NewStore(databaseMock)

	rows := sqlmock.NewRows([]string{"id", "name", "description", "price", "category"}).
		AddRow(1, "Espresso", "Strong coffee", 3.50, "Coffee").
		AddRow(2, "Latte", "Milk coffee", 4.50, "Coffee")

	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT p.id, p.name, p.description, p.price, p.category FROM products p JOIN favorites f ON p.id = f.product_id WHERE f.user_id = $1")).
		WithArgs(1).
		WillReturnRows(rows)

	favorites, queryError := favoriteStore.GetUserFavorites(context.Background(), 1)
	assert.NoError(t, queryError)
	assert.Len(t, favorites, 2)
	assert.Equal(t, "Espresso", favorites[0].Name)
	assert.Equal(t, 3.50, favorites[0].Price)
	assert.Equal(t, "Latte", favorites[1].Name)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestStore_GetUserFavorites_Empty(t *testing.T) {
	databaseMock, sqlMock, mockError := sqlmock.New()
	if mockError != nil {
		t.Fatalf("failed to open sqlmock: %s", mockError)
	}
	defer databaseMock.Close()

	favoriteStore := NewStore(databaseMock)

	rows := sqlmock.NewRows([]string{"id", "name", "description", "price", "category"})

	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT p.id, p.name, p.description, p.price, p.category FROM products p JOIN favorites f ON p.id = f.product_id WHERE f.user_id = $1")).
		WithArgs(99).
		WillReturnRows(rows)

	favorites, queryError := favoriteStore.GetUserFavorites(context.Background(), 99)
	assert.NoError(t, queryError)
	assert.Empty(t, favorites)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
