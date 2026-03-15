package user

import (
	usermodel "coffeebase-api/internal/models/user"
	"context"
	"database/sql"
)

// --- Public ---

func (store *postgresStore) Create(context context.Context, user *usermodel.User) error {
	if user.Language == "" {
		user.Language = "es" 
	}
	if user.Role == "" {
		user.Role = usermodel.RoleCustomer
	}
	query := `INSERT INTO users (username, email, password, language, avatar_url, role) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	return store.databaseConnection.QueryRowContext(context, query, user.Username, user.Email, user.Password, user.Language, user.AvatarURL, user.Role).Scan(&user.ID, &user.CreatedAt)
}

func (store *postgresStore) GetByEmail(context context.Context, email string) (usermodel.User, error) {
	query := store.getBaseUserQuery() + ` WHERE LOWER(email) = LOWER($1)`
	row := store.databaseConnection.QueryRowContext(context, query, email)
	return store.scanUserRow(row)
}

func (store *postgresStore) GetByID(context context.Context, id int) (usermodel.User, error) {
	query := store.getBaseUserQuery() + ` WHERE id = $1`
	row := store.databaseConnection.QueryRowContext(context, query, id)
	return store.scanUserRow(row)
}

func (store *postgresStore) UpdateAvatar(context context.Context, id int, avatarURL string) error {
	query := `UPDATE users SET avatar_url = $1 WHERE id = $2`
	_, error := store.databaseConnection.ExecContext(context, query, avatarURL, id)
	return error
}

func (store *postgresStore) UpdateLanguage(context context.Context, id int, language string) error {
	query := `UPDATE users SET language = $1 WHERE id = $2`
	_, error := store.databaseConnection.ExecContext(context, query, language, id)
	return error
}

func (store *postgresStore) GetTotalCount(context context.Context) (int, error) {
	var count int
	error := store.databaseConnection.QueryRowContext(context, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, error
}

func (store *postgresStore) GetAll(context context.Context) ([]usermodel.User, error) {
	query := `
		SELECT 
			id, username, email, language, avatar_url, role, 
			total_orders_completed, total_spent, created_at,
			COALESCE(first_name, ''), COALESCE(last_name, ''), birthday 
		FROM users ORDER BY created_at DESC`
	rows, error := store.databaseConnection.QueryContext(context, query)
	if error != nil {
		return nil, error
	}
	defer rows.Close()

	var users []usermodel.User
	for rows.Next() {
		var user usermodel.User
		var birthday sql.NullTime
		if error := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Language, &user.AvatarURL, &user.Role, 
			&user.TotalOrdersCompleted, &user.TotalSpent, &user.CreatedAt,
			&user.FirstName, &user.LastName, &birthday,
		); error != nil {
			return nil, error
		}
		if birthday.Valid {
			user.Birthday = birthday.Time
		}
		users = append(users, user)
	}
	return users, nil
}

func (store *postgresStore) UpdateRole(context context.Context, id int, role usermodel.UserRole) error {
	query := `UPDATE users SET role = $1 WHERE id = $2`
	_, error := store.databaseConnection.ExecContext(context, query, role, id)
	return error
}

func (store *postgresStore) Update(context context.Context, user *usermodel.User) error {
	query := `UPDATE users SET first_name = $1, last_name = $2, birthday = $3, language = $4, username = $5 WHERE id = $6`
	_, error := store.databaseConnection.ExecContext(context, query, user.FirstName, user.LastName, user.Birthday, user.Language, user.Username, user.ID)
	return error
}

// --- Private ---

func (store *postgresStore) getBaseUserQuery() string {
	return `
		SELECT 
			id, username, email, password, 
			COALESCE(language, 'es'), 
			COALESCE(avatar_url, ''), 
			role, 
			total_orders_completed, total_spent, created_at,
			COALESCE(first_name, ''),
			COALESCE(last_name, ''),
			birthday
		FROM users`
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func (store *postgresStore) scanUserRow(row rowScanner) (usermodel.User, error) {
	var user usermodel.User
	var birthday sql.NullTime
	error := row.Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.Language, &user.AvatarURL, &user.Role, 
		&user.TotalOrdersCompleted, &user.TotalSpent, &user.CreatedAt,
		&user.FirstName, &user.LastName, &birthday,
	)
	if birthday.Valid {
		user.Birthday = birthday.Time
	}
	return user, error
}
