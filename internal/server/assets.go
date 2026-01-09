// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package server

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"codeberg.org/oliverandrich/go-webapp-template/internal/appcontext"
)

// findAssets scans the static directory for hashed asset files.
func findAssets() (*appcontext.Assets, error) {
	assets := &appcontext.Assets{}

	// Find CSS file with hash
	cssMatches, err := filepath.Glob("static/css/styles.*.css")
	if err != nil {
		return nil, fmt.Errorf("failed to glob CSS files: %w", err)
	}

	if len(cssMatches) == 0 {
		slog.Warn("no hashed CSS file found, using fallback")
		assets.CSSPath = "/static/css/styles.css"
	} else {
		assets.CSSPath = "/" + strings.ReplaceAll(cssMatches[0], "\\", "/")
	}

	// Find htmx JS file with hash
	jsMatches, err := filepath.Glob("static/js/htmx.*.js")
	if err != nil {
		return nil, fmt.Errorf("failed to glob JS files: %w", err)
	}

	if len(jsMatches) == 0 {
		slog.Warn("no hashed htmx file found, using fallback")
		assets.JSPath = "/static/js/htmx.js"
	} else {
		assets.JSPath = "/" + strings.ReplaceAll(jsMatches[0], "\\", "/")
	}

	slog.Debug("assets discovered", "css", assets.CSSPath, "js", assets.JSPath)
	return assets, nil
}
