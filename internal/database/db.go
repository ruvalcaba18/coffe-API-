package database

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

// --- Public ---

func NewConnection() (*sql.DB, error) {
	connectionString := os.Getenv("DATABASE_URL")
	if connectionString == "" {
		host := os.Getenv("DB_HOST")
		if host == "" { host = getEnvOrDefault("PGHOST", "localhost") }

		port := os.Getenv("DB_PORT")
		if port == "" { port = os.Getenv("DBPORT") }
		if port == "" { port = getEnvOrDefault("PGPORT", "5432") }

		databaseUser := os.Getenv("DB_USER")
		if databaseUser == "" { databaseUser = getEnvOrDefault("PGUSER", "postgres") }

		password := os.Getenv("DB_PASSWORD")
		if password == "" { password = os.Getenv("PGPASSWORD") }

		databaseName := os.Getenv("DB_NAME")
		if databaseName == "" { databaseName = getEnvOrDefault("PGDATABASE", "coffeeshop") }

		if password != "" {
			connectionString = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
				databaseUser, password, host, port, databaseName)
		} else {
			connectionString = fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=disable",
				databaseUser, host, port, databaseName)
		}
	}

	database, connectionError := sql.Open("postgres", connectionString)
	if connectionError != nil {
		return nil, connectionError
	}

	if pingError := database.Ping(); pingError != nil {
		return nil, pingError
	}

	maxOpenConnections := getEnvOrDefaultInt("DB_MAX_OPEN_CONNS", 50)
	maxIdleConnections := getEnvOrDefaultInt("DB_MAX_IDLE_CONNS", 25)
	connectionMaxLifetime := getEnvOrDefaultDuration("DB_CONN_MAX_LIFETIME_MINUTES", 5*time.Minute)
	connectionMaxIdleTime := getEnvOrDefaultDuration("DB_CONN_MAX_IDLE_TIME_MINUTES", 3*time.Minute)

	database.SetMaxOpenConns(maxOpenConnections)
	database.SetMaxIdleConns(maxIdleConnections)
	database.SetConnMaxLifetime(connectionMaxLifetime)
	database.SetConnMaxIdleTime(connectionMaxIdleTime)

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

func getEnvOrDefaultInt(key string, defaultValue int) int {
	valueString := os.Getenv(key)
	if valueString == "" {
		return defaultValue
	}
	parsedValue, parseError := strconv.Atoi(valueString)
	if parseError != nil {
		return defaultValue
	}
	return parsedValue
}

func getEnvOrDefaultDuration(key string, defaultValue time.Duration) time.Duration {
	valueString := os.Getenv(key)
	if valueString == "" {
		return defaultValue
	}
	parsedMinutes, parseError := strconv.Atoi(valueString)
	if parseError != nil {
		return defaultValue
	}
	return time.Duration(parsedMinutes) * time.Minute
}
