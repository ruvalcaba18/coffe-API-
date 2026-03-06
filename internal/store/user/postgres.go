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
	if u.Role == "" {
		u.Role = "customer"
	}
	query := `INSERT INTO users (username, email, password, language, avatar_url, role) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	return s.db.QueryRow(query, u.Username, u.Email, u.Password, u.Language, u.AvatarURL, u.Role).Scan(&u.ID, &u.CreatedAt)
}

func (s *Store) GetByEmail(email string) (usermodel.User, error) {
	var u usermodel.User
	query := `SELECT id, username, email, password, language, avatar_url, role, created_at FROM users WHERE email = $1`
	err := s.db.QueryRow(query, email).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Language, &u.AvatarURL, &u.Role, &u.CreatedAt)
	return u, err
}

func (s *Store) GetByID(id int) (usermodel.User, error) {
	var u usermodel.User
	query := `SELECT id, username, email, password, language, avatar_url, role, created_at FROM users WHERE id = $1`
	err := s.db.QueryRow(query, id).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Language, &u.AvatarURL, &u.Role, &u.CreatedAt)
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

func (s *Store) GetTotalCount() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func (s *Store) GetAll() ([]usermodel.User, error) {
	rows, err := s.db.Query(`SELECT id, username, email, language, avatar_url, role, created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []usermodel.User
	for rows.Next() {
		var u usermodel.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Language, &u.AvatarURL, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *Store) UpdateRole(id int, role string) error {
	query := `UPDATE users SET role = $1 WHERE id = $2`
	_, err := s.db.Exec(query, role, id)
	return err
}

