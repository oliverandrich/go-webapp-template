// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindAssets_WithHashedFiles(t *testing.T) {
	// Save current working directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)

	// Create temp directory and change to it
	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	// Create directory structure
	require.NoError(t, os.MkdirAll("static/css", 0755))
	require.NoError(t, os.MkdirAll("static/js", 0755))

	// Create hashed asset files
	require.NoError(t, os.WriteFile("static/css/styles.abc123.css", []byte(""), 0644))
	require.NoError(t, os.WriteFile("static/js/htmx.def456.js", []byte(""), 0644))

	assets, err := findAssets()

	require.NoError(t, err)
	assert.Equal(t, "/static/css/styles.abc123.css", assets.CSSPath)
	assert.Equal(t, "/static/js/htmx.def456.js", assets.JSPath)
}

func TestFindAssets_Fallback(t *testing.T) {
	// Save current working directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)

	// Create temp directory and change to it
	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	// Create directory structure but NO hashed files
	require.NoError(t, os.MkdirAll("static/css", 0755))
	require.NoError(t, os.MkdirAll("static/js", 0755))

	assets, err := findAssets()

	require.NoError(t, err)
	assert.Equal(t, "/static/css/styles.css", assets.CSSPath)
	assert.Equal(t, "/static/js/htmx.js", assets.JSPath)
}

func TestFindAssets_PartialMatch(t *testing.T) {
	// Save current working directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)

	// Create temp directory and change to it
	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	// Create directory structure
	require.NoError(t, os.MkdirAll("static/css", 0755))
	require.NoError(t, os.MkdirAll("static/js", 0755))

	// Only CSS has a hashed file
	require.NoError(t, os.WriteFile("static/css/styles.xyz789.css", []byte(""), 0644))

	assets, err := findAssets()

	require.NoError(t, err)
	assert.Equal(t, "/static/css/styles.xyz789.css", assets.CSSPath)
	assert.Equal(t, "/static/js/htmx.js", assets.JSPath) // fallback
}

func TestFindAssets_MultipleMatches(t *testing.T) {
	// Save current working directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)

	// Create temp directory and change to it
	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	// Create directory structure
	require.NoError(t, os.MkdirAll("static/css", 0755))
	require.NoError(t, os.MkdirAll("static/js", 0755))

	// Multiple CSS files - filepath.Glob returns sorted results
	require.NoError(t, os.WriteFile("static/css/styles.aaa111.css", []byte(""), 0644))
	require.NoError(t, os.WriteFile("static/css/styles.zzz999.css", []byte(""), 0644))
	require.NoError(t, os.WriteFile("static/js/htmx.bbb222.js", []byte(""), 0644))

	assets, err := findAssets()

	require.NoError(t, err)
	// Should use first match (alphabetically sorted)
	assert.Equal(t, "/static/css/styles.aaa111.css", assets.CSSPath)
	assert.Equal(t, "/static/js/htmx.bbb222.js", assets.JSPath)
}

func TestFindAssets_NoDirectory(t *testing.T) {
	// Save current working directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)

	// Create temp directory and change to it
	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	// No static directory at all - Glob doesn't error, just returns empty
	assets, err := findAssets()

	require.NoError(t, err)
	assert.Equal(t, "/static/css/styles.css", assets.CSSPath)
	assert.Equal(t, "/static/js/htmx.js", assets.JSPath)
}

func TestFindAssets_PathSeparatorNormalization(t *testing.T) {
	// Save current working directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)

	// Create temp directory and change to it
	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	// Create directory structure
	require.NoError(t, os.MkdirAll(filepath.Join("static", "css"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join("static", "js"), 0755))

	// Create files using filepath.Join (OS-specific separators)
	require.NoError(t, os.WriteFile(filepath.Join("static", "css", "styles.hash123.css"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join("static", "js", "htmx.hash456.js"), []byte(""), 0644))

	assets, err := findAssets()

	require.NoError(t, err)
	// Paths should always use forward slashes (URL format)
	assert.Equal(t, "/static/css/styles.hash123.css", assets.CSSPath)
	assert.Equal(t, "/static/js/htmx.hash456.js", assets.JSPath)
}
