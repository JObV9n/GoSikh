package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Config holds SQLite connection configuration
type Config struct {
	Path            string        // Path to SQLite database file
	MaxOpenConns    int           // Maximum number of open connections
	MaxIdleConns    int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum lifetime of a connection
}

// DB wraps the sql.DB connection
type DB struct {
	*sql.DB
}

// New creates a new SQLite database connection
func New(config Config) (*DB, error) {
	// Set defaults if not provided
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = 10
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 10
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = 0
	}

	// Open database connection
	dsn := config.Path
	if strings.Contains(dsn, "?") {
		dsn += "&_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL&_synchronous=NORMAL"
	} else {
		dsn += "?_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL&_synchronous=NORMAL"
	}

	// Keep lock-free reads fast and avoid temp-disk overhead where possible.
	dsn += "&_cache_size=" + url.QueryEscape("-20000") + "&_temp_store=MEMORY"

	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{sqlDB}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db == nil || db.DB == nil {
		return nil
	}
	return db.DB.Close()
}

// HealthCheck verifies the database connection is healthy
func (db *DB) HealthCheck(ctx context.Context) error {
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	return nil
}
