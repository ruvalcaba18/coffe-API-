package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
)

func RunMigrations(database *sql.DB) error {
	_, errorResult := database.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if errorResult != nil {
		return fmt.Errorf("could not create migrations table: %v", errorResult)
	}

	var migrationFiles []string
	errorResult = filepath.Walk("internal/database/migrations", func(path string, info os.FileInfo, walkError error) error {
		if walkError != nil {
			return walkError
		}
		if !info.IsDir() && filepath.Ext(path) == ".sql" {
			migrationFiles = append(migrationFiles, path)
		}
		return nil
	})
	if errorResult != nil {
		return fmt.Errorf("could not read migrations directory: %v", errorResult)
	}
	sort.Strings(migrationFiles)

	for _, fullPath := range migrationFiles {
		filename := filepath.Base(fullPath)
		var exists bool
		errorResult := database.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", filename).Scan(&exists)
		if errorResult != nil {
			return errorResult
		}

		if exists {
			continue
		}

		log.Printf("Applying migration: %s", filename)
		content, errorResult := os.ReadFile(fullPath)
		if errorResult != nil {
			return errorResult
		}

		transaction, errorResult := database.Begin()
		if errorResult != nil {
			return errorResult
		}

		if _, errorResult := transaction.Exec(string(content)); errorResult != nil {
			transaction.Rollback()
			return fmt.Errorf("errorResult in %s: %v", filename, errorResult)
		}

		if _, errorResult := transaction.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", filename); errorResult != nil {
			transaction.Rollback()
			return errorResult
		}

		if errorResult := transaction.Commit(); errorResult != nil {
			return errorResult
		}
		log.Printf("Successfully applied: %s", filename)
	}

	return nil
}
