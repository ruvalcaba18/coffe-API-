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
	databaseMock, sqlMock, error := sqlmock.New()
	if error != nil {
		t.Fatalf("failed to open sqlmock: %s", error)
	}
	defer databaseMock.Close()

	userStore := userstore.NewStore(databaseMock)
	authHandler := &AuthHandler{userStore: userStore}

	email := "test@example.com"
	password := "correct-password"
	hashedPassword, _ := auth.HashPassword(password)

	loginBody := map[string]string{
		"email":    email,
		"password": password,
	}
	bodyData, _ := json.Marshal(loginBody)

	now := time.Now()
rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "language", "avatar_url", "role", "total_orders_completed", "total_spent", "created_at", "first_name", "last_name", "birthday"}).
		AddRow(1, "testuser", email, hashedPassword, "en", "", "customer", 0, 0, now, "", "", nil)

	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, password, COALESCE(language, 'es'), COALESCE(avatar_url, ''), role, total_orders_completed, total_spent, created_at, COALESCE(first_name, ''), COALESCE(last_name, ''), birthday FROM users WHERE LOWER(email) = LOWER($1)")).
		WithArgs(email).
		WillReturnRows(rows)

	httpRequest, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(bodyData))
	responseRecorder := httptest.NewRecorder()

	authHandler.Login(responseRecorder, httpRequest)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d. Body: %s", responseRecorder.Code, responseRecorder.Body.String())
	}
	
	var response map[string]string
	json.Unmarshal(responseRecorder.Body.Bytes(), &response)
	assert.NotEmpty(t, response["token"])

	cookies := responseRecorder.Result().Cookies()
	var authCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "auth-token" {
			authCookie = cookie
			break
		}
	}
	if authCookie == nil {
		t.Fatal("auth-token cookie not found")
	}
	assert.Equal(t, response["token"], authCookie.Value)

	if error := sqlMock.ExpectationsWereMet(); error != nil {
		t.Errorf("there were unfulfilled expectations: %s", error)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	databaseMock, sqlMock, error := sqlmock.New()
	if error != nil {
		t.Fatalf("failed to open sqlmock: %s", error)
	}
	defer databaseMock.Close()

	userStore := userstore.NewStore(databaseMock)
	authHandler := &AuthHandler{userStore: userStore}

	email := "test@example.com"
	password := "wrong-password"

	loginBody := map[string]string{
		"email":    email,
		"password": password,
	}
	bodyData, _ := json.Marshal(loginBody)

	hashedPassword, _ := auth.HashPassword("correct-password")
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "language", "avatar_url", "role", "total_orders_completed", "total_spent", "created_at", "first_name", "last_name", "birthday"}).
		AddRow(1, "testuser", email, hashedPassword, "en", "", "customer", 0, 0, now, "", "", nil)

	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, password, COALESCE(language, 'es'), COALESCE(avatar_url, ''), role, total_orders_completed, total_spent, created_at, COALESCE(first_name, ''), COALESCE(last_name, ''), birthday FROM users WHERE LOWER(email) = LOWER($1)")).
		WithArgs(email).
		WillReturnRows(rows)

	httpRequest, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(bodyData))
	responseRecorder := httptest.NewRecorder()

	authHandler.Login(responseRecorder, httpRequest)

	assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

	if error := sqlMock.ExpectationsWereMet(); error != nil {
		t.Errorf("there were unfulfilled expectations: %s", error)
	}
}
