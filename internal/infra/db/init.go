package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"GoT0Emergency/internal/pkg/log"
	"GoT0Emergency/internal/pkg/path"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

//go:embed schema.sql
var schema string

func Init() error {
	dbPath := path.GetDBPath()
	log.Info("Initializing database", "path", dbPath)

	// Ensure db directory exists (should be handled by path.Init but good to double check or if called independently)
	// path.Init() guarantees it.

	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	DB.SetMaxOpenConns(1) // SQLite works best with 1 writer, but WAL allows multiple readers.
	// Actually with WAL, we can have concurrent readers.
	// Mattn sqlite3 driver handles concurrency well in WAL mode.
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return migrate()
}

func migrate() error {
	_, err := DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
