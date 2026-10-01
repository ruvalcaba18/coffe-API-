package token

import (
	"context"
	"database/sql"
	"time"

	tokenmodel "coffeebase-api/internal/models/token"

	"github.com/google/uuid"
)

// Store define las operaciones de persistencia para refresh tokens.
type Store interface {
	// Create persiste un nuevo refresh token.
	Create(ctx context.Context, token *tokenmodel.RefreshToken) error

	// GetByHash busca un refresh token por el hash HMAC-SHA256 del token raw.
	GetByHash(ctx context.Context, hash string) (*tokenmodel.RefreshToken, error)

	// MarkUsed marca un token específico como ya utilizado (rotado).
	MarkUsed(ctx context.Context, id uuid.UUID) error

	// InvalidateFamily invalida todos los tokens de una familia de sesión.
	// Se llama cuando se detecta reutilización de un token ya rotado (posible robo).
	InvalidateFamily(ctx context.Context, familyID uuid.UUID) error

	// DeleteExpired elimina tokens expirados (limpieza periódica).
	DeleteExpired(ctx context.Context) error
}

type postgresStore struct {
	db *sql.DB
}

// --- Public ---

func NewStore(db *sql.DB) Store {
	return &postgresStore{db: db}
}

// --- Private (implementación) ---

func (s *postgresStore) Create(ctx context.Context, token *tokenmodel.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, family_id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`
	return s.db.QueryRowContext(ctx, query,
		token.ID, token.FamilyID, token.UserID, token.TokenHash, token.ExpiresAt,
	).Scan(&token.CreatedAt)
}

func (s *postgresStore) GetByHash(ctx context.Context, hash string) (*tokenmodel.RefreshToken, error) {
	query := `
		SELECT id, family_id, user_id, token_hash, expires_at, used, created_at
		FROM refresh_tokens
		WHERE token_hash = $1`
	row := s.db.QueryRowContext(ctx, query, hash)

	var t tokenmodel.RefreshToken
	err := row.Scan(&t.ID, &t.FamilyID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.Used, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *postgresStore) MarkUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET used = TRUE WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

func (s *postgresStore) InvalidateFamily(ctx context.Context, familyID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET used = TRUE WHERE family_id = $1`
	_, err := s.db.ExecContext(ctx, query, familyID)
	return err
}

func (s *postgresStore) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < $1`
	_, err := s.db.ExecContext(ctx, query, time.Now())
	return err
}
