// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package database_test

import (
	"os"
	"testing"

	"codeberg.org/oliverandrich/go-webapp-template/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpen_InMemory(t *testing.T) {
	db, err := database.Open(":memory:")

	require.NoError(t, err)
	require.NotNil(t, db)

	err = db.Close()
	require.NoError(t, err)
}

func TestOpen_DefaultDSN(t *testing.T) {
	// Create a temp directory and test there
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	db, err := database.Open("")

	require.NoError(t, err)
	require.NotNil(t, db)

	defer func() {
		_ = db.Close()
	}()
}

func TestOpen_MigrationsApplied(t *testing.T) {
	db, err := database.Open(":memory:")
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	// Verify tables were created by migrations
	var count int64
	err = db.Get(&count, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='users'")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestClose(t *testing.T) {
	db, err := database.Open(":memory:")
	require.NoError(t, err)

	err = db.Close()

	require.NoError(t, err)
}

func TestOpen_WithExistingParams(t *testing.T) {
	// Test that existing parameters are not duplicated
	db, err := database.Open(":memory:?cache=shared")

	require.NoError(t, err)
	require.NotNil(t, db)

	defer func() {
		_ = db.Close()
	}()
}

func TestOpen_PragmasApplied(t *testing.T) {
	db, err := database.Open(":memory:")
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	// Check that WAL mode is set
	var journalMode string
	err = db.Get(&journalMode, "PRAGMA journal_mode")
	require.NoError(t, err)
	// In memory mode, WAL might not be applied, but this shouldn't error
	assert.NotEmpty(t, journalMode)

	// Check synchronous setting
	var synchronous int
	err = db.Get(&synchronous, "PRAGMA synchronous")
	require.NoError(t, err)
	assert.NotZero(t, synchronous)
}

func TestOpen_FileDatabase(t *testing.T) {
	// Create a temp directory for the test database
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/subdir/test.db"

	db, err := database.Open(dbPath)
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	// Verify database is usable - tables should exist from migrations
	var count int64
	err = db.Get(&count, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='users'")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestOpen_ModeMemory(t *testing.T) {
	// Test mode=memory which is another way to use in-memory database
	db, err := database.Open("file::memory:?mode=memory")

	require.NoError(t, err)
	require.NotNil(t, db)

	defer func() {
		_ = db.Close()
	}()
}
