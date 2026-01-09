// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codeberg.org/oliverandrich/go-webapp-template/internal/config"
	"codeberg.org/oliverandrich/go-webapp-template/internal/database"
	"codeberg.org/oliverandrich/go-webapp-template/internal/handlers"
	"codeberg.org/oliverandrich/go-webapp-template/internal/i18n"
	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"codeberg.org/oliverandrich/go-webapp-template/internal/repository"
	"github.com/labstack/echo/v4"
	"github.com/lmittmann/tint"
	"github.com/urfave/cli/v3"
)

// Run starts the server with the given CLI command.
func Run(ctx context.Context, cmd *cli.Command) error {
	cfg := config.NewFromCLI(cmd)
	setupLogger(cfg.Log.Level, cfg.Log.Format)

	slog.Info("starting server",
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
		"base_url", cfg.Server.BaseURL,
	)

	// Database
	db, err := database.Open(cfg.Database.DSN)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() {
		if closeErr := database.Close(db); closeErr != nil {
			slog.Error("failed to close database", "error", closeErr)
		}
	}()

	// Migrations
	if migrateErr := database.Migrate(db, models.AllModels()...); migrateErr != nil {
		return fmt.Errorf("failed to migrate database: %w", migrateErr)
	}

	// i18n
	if initErr := i18n.Init(); initErr != nil {
		return fmt.Errorf("failed to init i18n: %w", initErr)
	}

	// Repository
	repo := repository.New(db)

	// Echo
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Assets
	assets, err := findAssets()
	if err != nil {
		return fmt.Errorf("failed to find assets: %w", err)
	}

	// Middleware
	setupMiddleware(e, cfg, assets)

	// Routes
	setupRoutes(e, repo)

	// Start server
	return startWithGracefulShutdown(e, cfg)
}

func setupRoutes(e *echo.Echo, repo *repository.Repository) {
	h := handlers.New(repo)

	// Static files
	e.Static("/static", "static")

	// Routes
	e.GET("/health", h.Health)
	e.GET("/", h.Home)
}

func startWithGracefulShutdown(e *echo.Echo, cfg *config.Config) error {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	go func() {
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	slog.Info("server stopped")
	return nil
}

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
