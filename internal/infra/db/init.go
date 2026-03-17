// Package db provides database initialization and connection management
// Uses SQLite3 with WAL (Write-Ahead Logging) mode for concurrent read support
package db

import (
	"database/sql"
	_ "embed"
	"time"

	"GoT0Emergency/internal/pkg/log"
	"GoT0Emergency/internal/pkg/path"

	_ "github.com/mattn/go-sqlite3"
)

// DB is the global database connection handle
var DB *sql.DB

// schema contains the embedded SQL schema definition from schema.sql
//go:embed schema.sql
var schema string

// Init initializes the SQLite database connection
// Sets up WAL mode and connection pooling for optimal performance
// Returns: error if database initialization fails
func Init() error {
	dbPath := path.GetDBPath()
	log.Info("Initializing database", "path", dbPath)

	// Ensure db directory exists (should be handled by path.Init but good to double check or if called independently)
	// path.Init() guarantees it.

	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return log.Errorf("failed to open database: %w", err)
	}

	DB.SetMaxOpenConns(1) // SQLite works best with 1 writer, but WAL allows multiple readers.
	// Actually with WAL, we can have concurrent readers.
	// Mattn sqlite3 driver handles concurrency well in WAL mode.
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	if err = DB.Ping(); err != nil {
		return log.Errorf("failed to ping database: %w", err)
	}

	return migrate()
}

// migrate executes the schema definition to create/update tables
// Returns: error if schema execution fails
func migrate() error {
	_, err := DB.Exec(schema)
	if err != nil {
		return log.Errorf("failed to execute schema: %w", err)
	}
	return nil
}

// Close closes the database connection
// Should be called on application shutdown
func Close() {
	if DB != nil {
		DB.Close()
	}
}
