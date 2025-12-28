// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"codeberg.org/oliverandrich/go-webapp-template/internal/models"

	_ "modernc.org/sqlite" // Pure-Go SQLite driver
)

// Open opens a GORM database connection with optimized SQLite settings.
func Open(dsn string) (*gorm.DB, error) {
	// Handle empty DSN
	if dsn == "" {
		dsn = "./data/app.db"
	}

	// Extract path for directory creation (handle query parameters)
	path := dsn
	if idx := strings.Index(dsn, "?"); idx != -1 {
		path = dsn[:idx]
	}

	// Ensure directory exists for file-based databases
	if path != ":memory:" && !strings.HasPrefix(path, ":memory:") {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	// Add default connection parameters if not already specified
	defaults := map[string]string{
		"_txlock":       "immediate",
		"_busy_timeout": "5000",
		"_foreign_keys": "on",
	}

	separator := "?"
	if strings.Contains(dsn, "?") {
		separator = "&"
	}

	for key, value := range defaults {
		if !strings.Contains(dsn, key+"=") {
			dsn += separator + key + "=" + value
			separator = "&"
		}
	}

	// Open sql.DB with modernc.org/sqlite driver
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create GORM connection using the existing sql.DB
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to create GORM connection: %w", err)
	}

	// Configure SQLite for better multi-user performance
	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA temp_store = MEMORY",
		"PRAGMA mmap_size = 134217728",
		"PRAGMA journal_size_limit = 27103364",
		"PRAGMA cache_size = 2000",
	}

	for _, pragma := range pragmas {
		if err := db.Exec(pragma).Error; err != nil {
			sqlDB.Close()
			return nil, fmt.Errorf("failed to set pragma: %w", err)
		}
	}

	return db, nil
}

// Migrate runs GORM AutoMigrate for all models.
// Note: The sessions table is managed by gormstore automatically.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(models.AllModels()...); err != nil {
		return fmt.Errorf("failed to auto-migrate: %w", err)
	}
	return nil
}
