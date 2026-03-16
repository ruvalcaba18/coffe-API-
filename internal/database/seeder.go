package database

import (
	"coffeebase-api/internal/auth"
	"coffeebase-api/internal/models/user"
	"database/sql"
	"strings"
)

// --- Public ---

func SeedDatabase(database *sql.DB) error {
	users := []struct {
		Username  string
		Email     string
		Password  string
		Role      user.UserRole
		FirstName string
		LastName  string
	}{
		{
			Username:  "Jael Admin",
			Email:     "jael.ruvalcaba@uabc.edu.mx",
			Password:  "User123!",
			Role:      user.RoleSuperAdmin,
			FirstName: "Jael",
			LastName:  "Admin",
		},
		{
			Username: "ian",
			Email:    "ian@gmail.com",
			Password: "user123!",
			Role:     user.RoleSuperAdmin,
			FirstName: "Ian",
			LastName:  "Master",
		},
		{
			Username:  "tester_user",
			Email:     "tester@example.com",
			Password:  "tester123!",
			Role:      user.RoleCustomer,
			FirstName: "John",
			LastName:  "Tester",
		},
		{
			Username:  "admin_test",
			Email:     "admin@example.com",
			Password:  "admin123!",
			Role:      user.RoleAdmin,
			FirstName: "Admin",
			LastName:  "Test",
		},
		{
			Username:  "barista_test",
			Email:     "barista@example.com",
			Password:  "barista123!",
			Role:      user.RoleBarista,
			FirstName: "Barista",
			LastName:  "Staff",
		},
	}

	for _, userEntry := range users {
		userEntry.Email = strings.ToLower(strings.TrimSpace(userEntry.Email))
		var id int
		var exists bool
		error := database.QueryRow("SELECT id, EXISTS(SELECT 1 FROM users WHERE LOWER(email) = LOWER($1)) FROM users WHERE LOWER(email) = LOWER($1) GROUP BY id", userEntry.Email).Scan(&id, &exists)
		
		if error != nil {
			id = 0
			exists = false
		}

		hashedPassword, error := auth.HashPassword(userEntry.Password)
		if error != nil {
			return error
		}

		if !exists {
			error = database.QueryRow(`
				INSERT INTO users (username, email, password, language, role, first_name, last_name, birthday) 
				VALUES ($1, $2, $3, 'es', $4, $5, $6, '1990-01-01') RETURNING id`,
				userEntry.Username, userEntry.Email, hashedPassword, userEntry.Role, userEntry.FirstName, userEntry.LastName,
			).Scan(&id)
			if error != nil {
				return error
			}
		} else {
			error = executeStatement(database, `
				UPDATE users SET role = $1, password = $2, username = $3, first_name = $4, last_name = $5
				WHERE email = $6`,
				userEntry.Role, hashedPassword, userEntry.Username, userEntry.FirstName, userEntry.LastName, userEntry.Email,
			)
			if error != nil {
				return error
			}
		}

		if userEntry.Username == "tester_user" {
			seedTesterDetails(database, id)
		}
	}

	return nil
}

// --- Private ---

func seedTesterDetails(database *sql.DB, userID int) {
	executeStatement(database, "UPDATE users SET total_orders_completed = 14, total_spent = 250.50 WHERE id = $1", userID)

	rows, error := database.Query("SELECT id FROM products LIMIT 5")
	if error != nil {
		return
	}
	defer rows.Close()

	var productIDs []int
	for rows.Next() {
		var id int
		if error := rows.Scan(&id); error == nil {
			productIDs = append(productIDs, id)
		}
	}

	if len(productIDs) == 0 {
		return
	}

	for _, productID := range productIDs[:2] {
		executeStatement(database, "INSERT INTO favorites (user_id, product_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, productID)
	}

	comments := []string{"¡Excelente sabor!", "Muy recomendado", "Ambiente increíble", "El mejor café de la ciudad"}
	for i, productID := range productIDs {
		rating := 4 + (i % 2)
		executeStatement(database, "INSERT INTO reviews (user_id, product_id, rating, comment) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING",
			userID, productID, rating, comments[i%len(comments)])
	}

	paymentMethodsQuery := `INSERT INTO payment_methods (user_id, last4, expiry, brand, holder, is_default) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`
	executeStatement(database, paymentMethodsQuery, userID, "4242", "12/26", "Visa", "John Tester", true)
	executeStatement(database, paymentMethodsQuery, userID, "5555", "09/25", "Mastercard", "John Tester", false)
}

func executeStatement(database *sql.DB, query string, args ...interface{}) error {
	_, error := database.Exec(query, args...)
	return error
}
