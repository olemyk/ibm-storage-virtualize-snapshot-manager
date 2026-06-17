package db

import (
	"database/sql"
	_ "embed"
	"fmt"

	"github.com/ibm-storage-virtualize-snapshot-manager/internal/config"
	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

//go:embed schema.sql
var schemaSQLite string

//go:embed schema_postgres.sql
var schemaPostgres string

// DB wraps the database connection
type DB struct {
	*sql.DB
	dbType string
}

// New creates a new database connection
func New(cfg *config.DatabaseConfig) (*DB, error) {
	var db *sql.DB
	var err error

	switch cfg.Type {
	case "sqlite":
		// SQLite connection settings
		const (
			sqliteJournalMode  = "WAL" // Write-Ahead Logging for better concurrency
			sqliteBusyTimeout  = 5000  // 5 seconds timeout for busy database
			sqliteMaxOpenConns = 10    // Allow multiple readers, single writer
			sqliteMaxIdleConns = 5     // Keep idle connections for reuse
		)

		// Add WAL mode and busy timeout to connection string
		connStr := fmt.Sprintf("%s?_journal_mode=%s&_busy_timeout=%d", cfg.Path, sqliteJournalMode, sqliteBusyTimeout)
		db, err = sql.Open("sqlite3", connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite database: %w", err)
		}
		// Enable foreign keys for SQLite
		if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
			return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
		}
		// Set connection pool settings - WAL mode allows multiple readers
		db.SetMaxOpenConns(sqliteMaxOpenConns)
		db.SetMaxIdleConns(sqliteMaxIdleConns)
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open postgres database: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db, dbType: cfg.Type}, nil
}

// Initialize creates database schema
func (db *DB) Initialize() error {
	var schemaSQL string

	switch db.dbType {
	case "sqlite":
		schemaSQL = schemaSQLite
	case "postgres":
		schemaSQL = schemaPostgres
	default:
		return fmt.Errorf("unsupported database type: %s", db.dbType)
	}

	_, err := db.Exec(schemaSQL)
	if err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// QueryRow wraps sql.DB.QueryRow and converts placeholders for PostgreSQL
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	if db.dbType == "postgres" {
		query = convertPlaceholders(query)
	}
	return db.DB.QueryRow(query, args...)
}

// Query wraps sql.DB.Query and converts placeholders for PostgreSQL
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if db.dbType == "postgres" {
		query = convertPlaceholders(query)
	}
	return db.DB.Query(query, args...)
}

// Exec wraps sql.DB.Exec and converts placeholders for PostgreSQL
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	if db.dbType == "postgres" {
		query = convertPlaceholders(query)
	}
	return db.DB.Exec(query, args...)
}

// convertPlaceholders converts ? placeholders to $1, $2, $3 for PostgreSQL
func convertPlaceholders(query string) string {
	var result []rune
	paramNum := 1
	inString := false
	escapeNext := false

	for _, ch := range query {
		if escapeNext {
			result = append(result, ch)
			escapeNext = false
			continue
		}

		if ch == '\\' {
			result = append(result, ch)
			escapeNext = true
			continue
		}

		if ch == '\'' {
			inString = !inString
			result = append(result, ch)
			continue
		}

		if ch == '?' && !inString {
			// Replace ? with $N
			result = append(result, []rune(fmt.Sprintf("$%d", paramNum))...)
			paramNum++
		} else {
			result = append(result, ch)
		}
	}

	return string(result)
}

//
