// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package database

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vinovest/sqlx"
	_ "modernc.org/sqlite" // Pure-Go SQLite driver
)

// Open creates a new database connection with optimized SQLite settings.
func Open(dsn string) (*sqlx.DB, error) {
	if dsn == "" {
		dsn = "./data/app.db"
	}

	// Create directory for file-based databases
	if !strings.HasPrefix(dsn, ":memory:") && !strings.Contains(dsn, "mode=memory") {
		dir := filepath.Dir(dsn)
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return nil, err
		}
	}

	// Add default SQLite parameters if not present
	dsn = addDefaultParams(dsn)

	// Open database with modernc.org/sqlite (pure-Go, CGO-free)
	conn, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(time.Hour)

	// Configure SQLite for better performance
	ctx := context.Background()
	if err := configureSQLite(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, err
	}

	// Run migrations
	if err := RunMigrations(conn.DB); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return conn, nil
}

// addDefaultParams adds recommended SQLite parameters if not already present.
func addDefaultParams(dsn string) string {
	defaults := map[string]string{
		"_txlock":       "immediate",
		"_busy_timeout": "5000",
		"_foreign_keys": "on",
	}

	for key, value := range defaults {
		if !strings.Contains(dsn, key) {
			separator := "?"
			if strings.Contains(dsn, "?") {
				separator = "&"
			}
			dsn += separator + key + "=" + value
		}
	}

	return dsn
}

// configureSQLite sets PRAGMAs for optimal performance.
func configureSQLite(ctx context.Context, db *sqlx.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA temp_store = MEMORY",
		"PRAGMA mmap_size = 134217728",
		"PRAGMA journal_size_limit = 27103364",
		"PRAGMA cache_size = 2000",
	}

	for _, pragma := range pragmas {
		if _, err := db.ExecContext(ctx, pragma); err != nil {
			return err
		}
	}

	return nil
}
