// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"codeberg.org/oliverandrich/go-webapp-template/internal/config"
	"codeberg.org/oliverandrich/go-webapp-template/internal/database"
	"codeberg.org/oliverandrich/go-webapp-template/internal/i18n"
	"codeberg.org/oliverandrich/go-webapp-template/internal/repository"
	"codeberg.org/oliverandrich/go-webapp-template/internal/services/auth"
	"codeberg.org/oliverandrich/go-webapp-template/internal/services/session"
	"github.com/go-chi/chi/v5"
	"github.com/urfave/cli/v3"
)

// setupLogger creates a structured logger based on configuration
func setupLogger(level, format string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func runServer(ctx context.Context, cmd *cli.Command) error {
	// Build configuration from CLI context
	cfg := config.NewFromCLI(cmd)

	// Setup structured logging
	logger := setupLogger(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(logger)

	// Open SQLite database
	db, err := database.Open(cfg.Database.DSN)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	defer sqlDB.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize i18n
	if err := i18n.Init(); err != nil {
		return fmt.Errorf("failed to initialize i18n: %w", err)
	}

	// Create repository
	repo := repository.New(db)

	// Create services
	authService := auth.NewService(repo, &cfg.Auth)

	// Create SCS session manager with GORM store
	sessionDuration := time.Duration(cfg.Auth.SessionDuration) * time.Hour
	rememberMeDuration := time.Duration(cfg.Auth.RememberMeDuration) * time.Hour
	sessionManager, err := session.NewSessionManager(db, sessionDuration, rememberMeDuration, cfg.Auth.CookieName, cfg.Auth.CookieSecure)
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}

	// Setup chi router
	r := chi.NewRouter()

	// Configure routes
	deps := &routerDeps{
		cfg:            cfg,
		repo:           repo,
		sessionManager: sessionManager,
		authService:    authService,
		logger:         logger,
	}
	setupRoutes(r, deps)

	// Log server configuration
	logger.Info("server_config",
		"registration_mode", cfg.Auth.RegistrationMode,
		"database", cfg.Database.DSN,
		"log_level", cfg.Log.Level,
	)

	// Start HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("server_start", "addr", addr)
	return http.ListenAndServe(addr, r)
}
