package user

import (
	usermodel "coffeebase-api/internal/models/user"
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
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", error)
	}
	defer databaseMock.Close()

	userStore := NewStore(databaseMock)

	userInstance := &usermodel.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	query := `INSERT INTO users \(username, email, password, language, avatar_url, role\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\) RETURNING id, created_at`
	
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now)
	
	sqlMock.ExpectQuery(query).
		WithArgs(userInstance.Username, userInstance.Email, userInstance.Password, "es", "", usermodel.RoleCustomer).
		WillReturnRows(rows)

	error = userStore.Create(context.Background(), userInstance)
	assert.NoError(t, error)
	assert.Equal(t, 1, userInstance.ID)
	assert.Equal(t, now, userInstance.CreatedAt)
	assert.Equal(t, "es", userInstance.Language)
	assert.Equal(t, usermodel.RoleCustomer, userInstance.Role)

	if error := sqlMock.ExpectationsWereMet(); error != nil {
		t.Errorf("there were unfulfilled expectations: %s", error)
	}
}

func TestStore_GetByEmail(t *testing.T) {
	databaseMock, sqlMock, error := sqlmock.New()
	if error != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", error)
	}
	defer databaseMock.Close()

	userStore := NewStore(databaseMock)

	email := "test@example.com"
	now := time.Now()
	
	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password", "language", "avatar_url", "role", 
		"total_orders_completed", "total_spent", "created_at", "first_name", "last_name", "birthday",
	}).AddRow(1, "testuser", email, "hashed_password", "en", "http://avatar.com", usermodel.RoleAdmin, 5, 150.50, now, "John", "Doe", nil)

	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, password, COALESCE(language, 'es'), COALESCE(avatar_url, ''), role, total_orders_completed, total_spent, created_at, COALESCE(first_name, ''), COALESCE(last_name, ''), birthday FROM users WHERE LOWER(email) = LOWER($1)")).
		WithArgs(email).
		WillReturnRows(rows)

	userInstance, error := userStore.GetByEmail(context.Background(), email)
	assert.NoError(t, error)
	assert.Equal(t, 1, userInstance.ID)
	assert.Equal(t, "testuser", userInstance.Username)
	assert.Equal(t, email, userInstance.Email)
	assert.Equal(t, "en", userInstance.Language)
	assert.Equal(t, usermodel.RoleAdmin, userInstance.Role)
	assert.Equal(t, 5, userInstance.TotalOrdersCompleted)
	assert.Equal(t, 150.50, userInstance.TotalSpent)

	if error := sqlMock.ExpectationsWereMet(); error != nil {
		t.Errorf("there were unfulfilled expectations: %s", error)
	}
}

func TestStore_UpdateAvatar(t *testing.T) {
	databaseMock, sqlMock, error := sqlmock.New()
	if error != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", error)
	}
	defer databaseMock.Close()

	userStore := NewStore(databaseMock)

	userID := 1
	avatarURL := "http://new-avatar.com"

	sqlMock.ExpectExec(regexp.QuoteMeta("UPDATE users SET avatar_url = $1 WHERE id = $2")).
		WithArgs(avatarURL, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	error = userStore.UpdateAvatar(context.Background(), userID, avatarURL)
	assert.NoError(t, error)

	if error := sqlMock.ExpectationsWereMet(); error != nil {
		t.Errorf("there were unfulfilled expectations: %s", error)
	}
}
