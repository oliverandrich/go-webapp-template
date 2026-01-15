// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

//go:build dev

// Package assets provides static asset serving for development mode.
// In development, assets are served directly from the filesystem without hashing.
package assets

import (
	"net/http"
)

// CSSPath returns the path to the main CSS file (unhashed in dev mode).
func CSSPath() string {
	return "/static/dist/styles.css"
}

// JSPath returns the path to the bundled JS file (unhashed in dev mode).
func JSPath() string {
	return "/static/dist/app.js"
}

// FileServer returns an http.Handler that serves static files from the filesystem.
func FileServer() http.Handler {
	return http.FileServer(http.Dir("internal/assets/static"))
}
