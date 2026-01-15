// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

//go:build !dev

// Package assets provides embedded static assets with content-hashed filenames.
package assets

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
)

//go:embed esbuild-meta.json
var metaData []byte

//go:embed static
var staticFS embed.FS

// esbuildMeta represents the esbuild metafile format.
type esbuildMeta struct {
	Outputs map[string]struct{} `json:"outputs"`
}

var (
	cssPath string
	jsPath  string
)

func init() {
	// Defaults (development fallback)
	cssPath = "/static/css/styles.css"
	jsPath = "/static/js/app.js"

	if len(metaData) == 0 {
		slog.Debug("esbuild meta is empty, using fallback paths")
		return
	}

	var meta esbuildMeta
	if err := json.Unmarshal(metaData, &meta); err != nil {
		slog.Error("failed to parse esbuild meta", "error", err)
		return
	}

	// Extract hashed paths from outputs
	for outputPath := range meta.Outputs {
		// Convert file path to URL: internal/assets/static/dist/... → /static/dist/...
		if strings.Contains(outputPath, "/static/") {
			idx := strings.Index(outputPath, "/static/")
			urlPath := outputPath[idx:]

			if strings.HasSuffix(urlPath, ".css") {
				cssPath = urlPath
			} else if strings.HasSuffix(urlPath, ".js") {
				jsPath = urlPath
			}
		}
	}

	slog.Debug("loaded asset paths", "css", cssPath, "js", jsPath)
}

// CSSPath returns the path to the main CSS file.
func CSSPath() string {
	return cssPath
}

// JSPath returns the path to the bundled JS file.
func JSPath() string {
	return jsPath
}

// FileServer returns an http.Handler that serves embedded static files.
func FileServer() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic("failed to create sub filesystem: " + err.Error())
	}
	return http.FileServer(http.FS(sub))
}
