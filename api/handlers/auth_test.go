package handlers

import (
	"bytes"
	"coffeebase-api/internal/auth"
	userstore "coffeebase-api/internal/store/user"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_Login_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	store := userstore.NewStore(db)
	handler := &AuthHandler{UserStore: store}

	email := "test@example.com"
	password := "correct-password"
	hashedPassword, _ := auth.HashPassword(password)

	loginBody := map[string]string{
		"email":    email,
		"password": password,
	}
	bodyData, _ := json.Marshal(loginBody)

	// Mock DB expectation for GetByEmail
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "language", "avatar_url", "role", "total_orders_completed", "total_spent", "created_at"}).
		AddRow(1, "testuser", email, hashedPassword, "en", "", "customer", 0, 0, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, password, COALESCE(language, 'es'), COALESCE(avatar_url, ''), role, total_orders_completed, total_spent, created_at FROM users WHERE LOWER(email) = LOWER($1)")).
		WithArgs(email).
		WillReturnRows(rows)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(bodyData))
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d. Body: %s", rr.Code, rr.Body.String())
	}
	
	var response map[string]string
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NotEmpty(t, response["token"])

	// Check cookie
	cookies := rr.Result().Cookies()
	var authCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "auth-token" {
			authCookie = c
			break
		}
	}
	if authCookie == nil {
		t.Fatal("auth-token cookie not found")
	}
	assert.Equal(t, response["token"], authCookie.Value)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	store := userstore.NewStore(db)
	handler := &AuthHandler{UserStore: store}

	email := "test@example.com"
	password := "wrong-password"

	loginBody := map[string]string{
		"email":    email,
		"password": password,
	}
	bodyData, _ := json.Marshal(loginBody)

	// Mock DB expectation for GetByEmail (user exists but password won't match)
	hashedPassword, _ := auth.HashPassword("correct-password")
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "language", "avatar_url", "role", "total_orders_completed", "total_spent", "created_at"}).
		AddRow(1, "testuser", email, hashedPassword, "en", "", "customer", 0, 0, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, password, COALESCE(language, 'es'), COALESCE(avatar_url, ''), role, total_orders_completed, total_spent, created_at FROM users WHERE LOWER(email) = LOWER($1)")).
		WithArgs(email).
		WillReturnRows(rows)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(bodyData))
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
