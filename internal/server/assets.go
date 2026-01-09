// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package server

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"codeberg.org/oliverandrich/go-webapp-template/internal/ctxkeys"
	"github.com/labstack/echo/v4"
)

// Assets holds paths to static assets.
type Assets struct {
	CSSPath string
}

// findAssets scans the static directory for hashed asset files.
func findAssets() (*Assets, error) {
	assets := &Assets{}

	// Find CSS file with hash
	matches, err := filepath.Glob("static/css/styles.*.css")
	if err != nil {
		return nil, fmt.Errorf("failed to glob CSS files: %w", err)
	}

	if len(matches) == 0 {
		slog.Warn("no hashed CSS file found, using fallback")
		assets.CSSPath = "/static/css/styles.css"
	} else {
		// Use the first match (should only be one)
		assets.CSSPath = "/" + strings.ReplaceAll(matches[0], "\\", "/")
	}

	slog.Debug("assets discovered", "css", assets.CSSPath)
	return assets, nil
}

// assetsToContext adds the asset paths to the request context.
func assetsToContext(assets *Assets) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := context.WithValue(c.Request().Context(), ctxkeys.CSSPath{}, assets.CSSPath)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
