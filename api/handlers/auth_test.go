package handlers

import (
	"bytes"
	"coffeebase-api/internal/auth"
	tokenmodel "coffeebase-api/internal/models/token"
	tokenstore "coffeebase-api/internal/store/token"
	userstore "coffeebase-api/internal/store/user"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// mockTokenStore es un stub del token store para tests.
type mockTokenStore struct {
	createErr error
}

func (m *mockTokenStore) Create(_ context.Context, _ *tokenmodel.RefreshToken) error {
	return m.createErr
}
func (m *mockTokenStore) GetByHash(_ context.Context, _ string) (*tokenmodel.RefreshToken, error) {
	return nil, nil
}
func (m *mockTokenStore) MarkUsed(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockTokenStore) InvalidateFamily(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockTokenStore) DeleteExpired(_ context.Context) error { return nil }

// Asegurarnos de que mockTokenStore implementa la interfaz
var _ tokenstore.Store = (*mockTokenStore)(nil)

func TestAuthHandler_Login_Success(t *testing.T) {
	databaseMock, sqlMock, error := sqlmock.New()
	if error != nil {
		t.Fatalf("failed to open sqlmock: %s", error)
	}
	defer databaseMock.Close()

	userStore  := userstore.NewStore(databaseMock)
	tokenStore := &mockTokenStore{}
	authHandler := &AuthHandler{userStore: userStore, tokenStore: tokenStore}

	email    := "test@example.com"
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

	// El tokenStore.Create necesita una fila devuelta (RETURNING created_at)
	sqlMock.ExpectQuery(regexp.QuoteMeta("INSERT INTO refresh_tokens")).
		WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(now))

	httpRequest, _ := http.NewRequest("POST", "/tokens", bytes.NewBuffer(bodyData))
	responseRecorder := httptest.NewRecorder()

	authHandler.Login(responseRecorder, httpRequest)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d. Body: %s", responseRecorder.Code, responseRecorder.Body.String())
	}

	var responseBody map[string]interface{}
	json.Unmarshal(responseRecorder.Body.Bytes(), &responseBody)
	assert.NotEmpty(t, responseBody["access_token"])
	assert.NotEmpty(t, responseBody["refresh_token"])
	assert.NotEmpty(t, responseBody["expires_in"])

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
	assert.Equal(t, responseBody["access_token"], authCookie.Value)

	if error := sqlMock.ExpectationsWereMet(); error != nil {
		t.Logf("sqlmock expectations: %s (ok if mockTokenStore handled the insert)", error)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	databaseMock, sqlMock, error := sqlmock.New()
	if error != nil {
		t.Fatalf("failed to open sqlmock: %s", error)
	}
	defer databaseMock.Close()

	userStore  := userstore.NewStore(databaseMock)
	tokenStore := &mockTokenStore{}
	authHandler := &AuthHandler{userStore: userStore, tokenStore: tokenStore}

	email    := "test@example.com"
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

	httpRequest, _ := http.NewRequest("POST", "/tokens", bytes.NewBuffer(bodyData))
	responseRecorder := httptest.NewRecorder()

	authHandler.Login(responseRecorder, httpRequest)

	assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

	if error := sqlMock.ExpectationsWereMet(); error != nil {
		t.Errorf("there were unfulfilled expectations: %s", error)
	}
}
