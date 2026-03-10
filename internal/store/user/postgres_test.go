package user

import (
	usermodel "coffeebase-api/internal/models/user"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := NewStore(db)

	u := &usermodel.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	query := `INSERT INTO users \(username, email, password, language, avatar_url, role\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\) RETURNING id, created_at`
	
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now)
	
	mock.ExpectQuery(query).
		WithArgs(u.Username, u.Email, u.Password, "es", "", "customer").
		WillReturnRows(rows)

	err = s.Create(u)
	assert.NoError(t, err)
	assert.Equal(t, 1, u.ID)
	assert.Equal(t, now, u.CreatedAt)
	assert.Equal(t, "es", u.Language)
	assert.Equal(t, "customer", u.Role)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestStore_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := NewStore(db)

	email := "test@example.com"
	now := time.Now()
	
	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password", "language", "avatar_url", "role", 
		"total_orders_completed", "total_spent", "created_at",
	}).AddRow(1, "testuser", email, "hashed_password", "en", "http://avatar.com", "admin", 5, 150.50, now)

	// Since we use multiline query, using regexp.QuoteMeta or a simplified regex
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, password, COALESCE(language, 'es'), COALESCE(avatar_url, ''), role, total_orders_completed, total_spent, created_at FROM users WHERE LOWER(email) = LOWER($1)")).
		WithArgs(email).
		WillReturnRows(rows)

	u, err := s.GetByEmail(email)
	assert.NoError(t, err)
	assert.Equal(t, 1, u.ID)
	assert.Equal(t, "testuser", u.Username)
	assert.Equal(t, email, u.Email)
	assert.Equal(t, "en", u.Language)
	assert.Equal(t, "admin", u.Role)
	assert.Equal(t, 5, u.TotalOrdersCompleted)
	assert.Equal(t, 150.50, u.TotalSpent)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestStore_UpdateAvatar(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := NewStore(db)

	id := 1
	avatarURL := "http://new-avatar.com"

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET avatar_url = $1 WHERE id = $2")).
		WithArgs(avatarURL, id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = s.UpdateAvatar(id, avatarURL)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
