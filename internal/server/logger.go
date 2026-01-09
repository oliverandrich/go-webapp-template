// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package server

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

// setupLogger configures the global slog logger.
func setupLogger(level, format string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	} else {
		handler = tint.NewHandler(os.Stdout, &tint.Options{Level: logLevel})
	}

	slog.SetDefault(slog.New(handler))
}
