package user

import (
	usermodel "coffeebase-api/internal/models/user"
	"database/sql"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(u *usermodel.User) error {
	if u.Language == "" {
		u.Language = "es" // Default
	}
	query := `INSERT INTO users (username, email, password, language, avatar_url) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	return s.db.QueryRow(query, u.Username, u.Email, u.Password, u.Language, u.AvatarURL).Scan(&u.ID, &u.CreatedAt)
}

func (s *Store) GetByEmail(email string) (usermodel.User, error) {
	var u usermodel.User
	query := `SELECT id, username, email, password, language, avatar_url, created_at FROM users WHERE email = $1`
	err := s.db.QueryRow(query, email).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Language, &u.AvatarURL, &u.CreatedAt)
	return u, err
}

func (s *Store) GetByID(id int) (usermodel.User, error) {
	var u usermodel.User
	query := `SELECT id, username, email, password, language, avatar_url, created_at FROM users WHERE id = $1`
	err := s.db.QueryRow(query, id).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Language, &u.AvatarURL, &u.CreatedAt)
	return u, err
}

func (s *Store) UpdateAvatar(id int, avatarURL string) error {
	query := `UPDATE users SET avatar_url = $1 WHERE id = $2`
	_, err := s.db.Exec(query, avatarURL, id)
	return err
}

func (s *Store) UpdateLanguage(id int, language string) error {
	query := `UPDATE users SET language = $1 WHERE id = $2`
	_, err := s.db.Exec(query, language, id)
	return err
}
