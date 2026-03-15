package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// --- Public ---

func NewConnection() (*sql.DB, error) {
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := os.Getenv("DB_PASSWORD")
	dbname := getEnvOrDefault("DB_NAME", "coffeeshop")

	auth := user
	if password != "" {
		auth = fmt.Sprintf("%s:%s", user, password)
	}

	connStr := fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=disable", auth, host, port, dbname)

	database, error := sql.Open("postgres", connStr)
	if error != nil {
		return nil, error
	}

	if error := database.Ping(); error != nil {
		return nil, error
	}

	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(25)
	database.SetConnMaxLifetime(5 * time.Minute)

	return database, nil
}

// --- Private ---

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
