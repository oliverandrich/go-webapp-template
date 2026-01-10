// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// Open creates a new database connection with optimized SQLite settings.
func Open(dsn string) (*gorm.DB, error) {
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

	// Open connection using modernc.org/sqlite (pure Go, no CGO)
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Use GORM's SQLite dialector with the existing connection
	db, err := gorm.Open(sqlite.New(sqlite.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	// Configure SQLite for better performance
	if configErr := configureSQLite(db); configErr != nil {
		return nil, configErr
	}

	return db, nil
}

// addDefaultParams adds recommended SQLite parameters if not already present.
// Uses modernc.org/sqlite pragma format: _pragma=name(value)
func addDefaultParams(dsn string) string {
	// modernc.org/sqlite uses _pragma=name(value) format
	defaults := []string{
		"_pragma=busy_timeout(5000)",
		"_pragma=foreign_keys(1)",
		"_txlock=immediate",
	}

	for _, param := range defaults {
		// Extract the pragma/param name for checking
		paramName := strings.Split(param, "=")[0]
		if !strings.Contains(dsn, paramName) {
			separator := "?"
			if strings.Contains(dsn, "?") {
				separator = "&"
			}
			dsn += separator + param
		}
	}

	return dsn
}

// configureSQLite sets PRAGMAs for optimal performance.
func configureSQLite(db *gorm.DB) error {
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
			return err
		}
	}

	return nil
}

// Migrate runs auto-migration for the given models.
func Migrate(db *gorm.DB, models ...any) error {
	return db.AutoMigrate(models...)
}

// Close closes the database connection.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
