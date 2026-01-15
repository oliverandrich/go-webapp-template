// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package server

import (
	"log/slog"

	"github.com/oliverandrich/go-webapp-template/internal/appcontext"
	"github.com/oliverandrich/go-webapp-template/internal/assets"
)

// findAssets returns asset paths from the embedded manifest.
func findAssets() *appcontext.Assets {
	a := &appcontext.Assets{
		CSSPath: assets.CSSPath(),
		JSPath:  assets.JSPath(),
	}
	slog.Debug("assets loaded", "css", a.CSSPath, "js", a.JSPath)
	return a
}
