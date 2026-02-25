package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func NewConnection() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if user == "" {
		user = "postgres"
	}
	if dbname == "" {
		dbname = "coffeeshop"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// 🛠 PUNTO 4: Configuración de Pool de Conexiones para Producción
	db.SetMaxOpenConns(25)                 // Límite máximo de conexiones abiertas al mismo tiempo
	db.SetMaxIdleConns(25)                 // Conexiones en espera (para no cerrar/abrir todo el tiempo)
	db.SetConnMaxLifetime(5 * time.Minute) // Expirar conexiones viejas para evitar leaks

	return db, nil
}
