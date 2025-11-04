package config

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{db: db}
	if err := database.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return database, nil
}

func (d *Database) migrate() error {
	// Create migrations table if it doesn't exist
	migrationTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		filename TEXT PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := d.db.Exec(migrationTableSQL); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrationDir := "migrations"
	files, err := ioutil.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		// Check if migration has already been applied
		var count int
		err := d.db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE filename = ?", file.Name()).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count > 0 {
			log.Printf("Migration already applied: %s", file.Name())
			continue
		}

		migrationPath := filepath.Join(migrationDir, file.Name())
		if err := d.runMigration(migrationPath, file.Name()); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", file.Name(), err)
		}

		log.Printf("Applied migration: %s", file.Name())
	}

	return nil
}

func (d *Database) runMigration(migrationPath, filename string) error {
	content, err := ioutil.ReadFile(migrationPath)
	if err != nil {
		return err
	}

	// Start transaction
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration
	_, err = tx.Exec(string(content))
	if err != nil {
		return err
	}

	// Record migration
	_, err = tx.Exec("INSERT INTO schema_migrations (filename) VALUES (?)", filename)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) GetDB() *sql.DB {
	return d.db
}