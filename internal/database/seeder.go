package database

import (
	"coffeebase-api/internal/auth"
	"coffeebase-api/internal/models/user"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// --- Public ---

func SeedDatabase(database *sql.DB) error {
	seedUsers := loadSeedUsersFromEnvironment()
	if len(seedUsers) == 0 {
		slog.Info("No seed users configured (set SEED_USERS env var to enable seeding)")
		return nil
	}

	for _, seedUser := range seedUsers {
		seedUser.Email = strings.ToLower(strings.TrimSpace(seedUser.Email))
		userID, exists := findExistingUser(database, seedUser.Email)

		hashedPassword, hashError := auth.HashPassword(seedUser.Password)
		if hashError != nil {
			return hashError
		}

		if !exists {
			userID, hashError = insertSeedUser(database, seedUser, hashedPassword)
			if hashError != nil {
				return hashError
			}
		} else {
			hashError = updateSeedUser(database, seedUser, hashedPassword)
			if hashError != nil {
				return hashError
			}
		}

		if seedUser.SeedProfile == "tester" {
			seedTesterDetails(database, userID)
		}

		if seedUser.SeedProfile == "demo" {
			seedDemoShowcaseDetails(database, userID)
		}
	}

	return nil
}

// --- Private ---

type seedUserEntry struct {
	Username    string
	Email       string
	Password    string
	Role        user.UserRole
	FirstName   string
	LastName    string
	SeedProfile string
}


func loadSeedUsersFromEnvironment() []seedUserEntry {
	seedUsersConfig := os.Getenv("SEED_USERS")
	if seedUsersConfig == "" {
		return nil
	}

	var seedUsers []seedUserEntry
	userEntries := strings.Split(seedUsersConfig, "|")

	for _, entry := range userEntries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		fields := strings.Split(entry, ":")
		if len(fields) < 6 {
			slog.Warn("Skipping malformed seed user entry (expected at least 6 fields separated by ':')", "entry", entry)
			continue
		}

		seedProfile := ""
		if len(fields) >= 7 {
			seedProfile = fields[6]
		}

		seedUsers = append(seedUsers, seedUserEntry{
			Username:    fields[0],
			Email:       fields[1],
			Password:    fields[2],
			Role:        user.UserRole(fields[3]),
			FirstName:   fields[4],
			LastName:    fields[5],
			SeedProfile: seedProfile,
		})
	}

	return seedUsers
}

func findExistingUser(database *sql.DB, email string) (int, bool) {
	var userID int
	var exists bool
	queryError := database.QueryRow(
		"SELECT id, EXISTS(SELECT 1 FROM users WHERE LOWER(email) = LOWER($1)) FROM users WHERE LOWER(email) = LOWER($1) GROUP BY id",
		email,
	).Scan(&userID, &exists)

	if queryError != nil {
		return 0, false
	}
	return userID, exists
}

func insertSeedUser(database *sql.DB, seedUser seedUserEntry, hashedPassword string) (int, error) {
	var userID int
	insertError := database.QueryRow(`
		INSERT INTO users (username, email, password, language, role, first_name, last_name, birthday)
		VALUES ($1, $2, $3, 'es', $4, $5, $6, '1990-01-01') RETURNING id`,
		seedUser.Username, seedUser.Email, hashedPassword, seedUser.Role, seedUser.FirstName, seedUser.LastName,
	).Scan(&userID)
	return userID, insertError
}

func updateSeedUser(database *sql.DB, seedUser seedUserEntry, hashedPassword string) error {
	return executeStatement(database, `
		UPDATE users SET role = $1, password = $2, username = $3, first_name = $4, last_name = $5
		WHERE email = $6`,
		seedUser.Role, hashedPassword, seedUser.Username, seedUser.FirstName, seedUser.LastName, seedUser.Email,
	)
}

func seedTesterDetails(database *sql.DB, userID int) {
	executeStatement(database, "UPDATE users SET total_orders_completed = 14, total_spent = 250.50 WHERE id = $1", userID)

	rows, queryError := database.Query("SELECT id FROM products LIMIT 5")
	if queryError != nil {
		return
	}
	defer rows.Close()

	var productIDs []int
	for rows.Next() {
		var productID int
		if scanError := rows.Scan(&productID); scanError == nil {
			productIDs = append(productIDs, productID)
		}
	}
	if rows.Err() != nil {
		return
	}

	if len(productIDs) == 0 {
		return
	}

	maxFavorites := 2
	if len(productIDs) < maxFavorites {
		maxFavorites = len(productIDs)
	}
	for _, productID := range productIDs[:maxFavorites] {
		executeStatement(database, "INSERT INTO favorites (user_id, product_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, productID)
	}

	comments := []string{"¡Excelente sabor!", "Muy recomendado", "Ambiente increíble", "El mejor café de la ciudad"}
	for index, productID := range productIDs {
		rating := 4 + (index % 2)
		executeStatement(database, "INSERT INTO reviews (user_id, product_id, rating, comment) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING",
			userID, productID, rating, comments[index%len(comments)])
	}

	paymentMethodsQuery := `INSERT INTO payment_methods (user_id, last4, expiry, brand, holder, is_default) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`
	executeStatement(database, paymentMethodsQuery, userID, "4242", "12/26", "Visa", "John Tester", true)
	executeStatement(database, paymentMethodsQuery, userID, "5555", "09/25", "Mastercard", "John Tester", false)
}

func seedDemoShowcaseDetails(database *sql.DB, userID int) {
	executeStatement(database, "UPDATE users SET total_orders_completed = 42, total_spent = 1250.75 WHERE id = $1", userID)

	rows, queryError := database.Query("SELECT id FROM products LIMIT 8")
	if queryError != nil {
		return
	}
	defer rows.Close()

	var productIDs []int
	for rows.Next() {
		var productID int
		if scanError := rows.Scan(&productID); scanError == nil {
			productIDs = append(productIDs, productID)
		}
	}
	if rows.Err() != nil {
		return
	}

	if len(productIDs) == 0 {
		return
	}

	maxFavorites := 4
	if len(productIDs) < maxFavorites {
		maxFavorites = len(productIDs)
	}
	for _, productID := range productIDs[:maxFavorites] {
		executeStatement(database, "INSERT INTO favorites (user_id, product_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, productID)
	}

	demoComments := []string{
		"Perfecto para el menú de temporada ☕",
		"Nuestros clientes lo piden todos los días",
		"Gran equilibrio entre acidez y cuerpo",
		"Ideal para recomendar a nuevos clientes",
		"La mejor opción para cold brew",
		"Excelente aroma, siempre consistente",
	}
	for index, productID := range productIDs {
		rating := 4 + (index % 2)
		executeStatement(database, "INSERT INTO reviews (user_id, product_id, rating, comment) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING",
			userID, productID, rating, demoComments[index%len(demoComments)])
	}

	paymentMethodsQuery := `INSERT INTO payment_methods (user_id, last4, expiry, brand, holder, is_default) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`
	executeStatement(database, paymentMethodsQuery, userID, "1234", "06/28", "Visa", "Demo LaBase", true)
	executeStatement(database, paymentMethodsQuery, userID, "9876", "03/27", "Mastercard", "Demo LaBase", false)
}

func executeStatement(database *sql.DB, query string, args ...interface{}) error {
	_, executionError := database.Exec(query, args...)
	if executionError != nil {
		slog.Error("Failed to execute seed statement", "error", executionError, "query", fmt.Sprintf("%.80s...", query))
	}
	return executionError
}
