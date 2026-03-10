package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
)

func RunMigrations(db *sql.DB) error {
	// 1. Create migrations table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("could not create migrations table: %v", err)
	}

	// 2. Read migration files recursively
	var migrationFiles []string
	err = filepath.Walk("internal/database/migrations", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".sql" {
			migrationFiles = append(migrationFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not read migrations directory: %v", err)
	}
	sort.Strings(migrationFiles)

	// 3. Apply each migration
	for _, fullPath := range migrationFiles {
		filename := filepath.Base(fullPath)
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", filename).Scan(&exists)
		if err != nil {
			return err
		}

		if exists {
			continue
		}

		log.Printf("Applying migration: %s", filename)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("error in %s: %v", filename, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", filename); err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
		log.Printf("Successfully applied: %s", filename)
	}

	return nil
}
