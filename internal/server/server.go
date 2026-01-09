// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
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
	// Setup TLS
	tlsResult, err := SetupTLS(cfg)
	if err != nil {
		return fmt.Errorf("TLS setup failed: %w", err)
	}

	// Channel for server errors
	errChan := make(chan error, 2)

	// HTTP redirect server for ACME mode
	var httpServer *http.Server

	switch tlsResult.Mode {
	case TLSModeOff:
		// Plain HTTP on configured port
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		go func() {
			slog.Info("Server running", "url", cfg.Server.BaseURL)
			if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errChan <- err
			}
		}()

	case TLSModeACME:
		// HTTPS on :443
		go func() {
			slog.Info("Server running", "url", cfg.Server.BaseURL)
			if err := startTLSServer(e, ":443", tlsResult.TLSConfig); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errChan <- err
			}
		}()

		// HTTP redirect server on :80
		httpServer = &http.Server{
			Addr:              ":80",
			Handler:           tlsResult.HTTPHandler,
			ReadHeaderTimeout: 10 * time.Second,
		}
		go func() {
			slog.Info("HTTP→HTTPS redirect active", "addr", ":80")
			if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errChan <- err
			}
		}()

	case TLSModeSelfSigned, TLSModeManual:
		// HTTPS on configured port
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		go func() {
			slog.Info("Server running", "url", cfg.Server.BaseURL)
			if err := startTLSServer(e, addr, tlsResult.TLSConfig); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errChan <- err
			}
		}()
	}

	// Wait for interrupt signal or error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		slog.Info("shutting down server")
	case err := <-errChan:
		slog.Error("server error", "error", err)
		return err
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown main server
	if err := e.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown main server", "error", err)
	}

	// Shutdown HTTP redirect server if running
	if httpServer != nil {
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("failed to shutdown HTTP redirect server", "error", err)
		}
	}

	slog.Info("server stopped")
	return nil
}

// startTLSServer starts the Echo server with a custom TLS configuration.
func startTLSServer(e *echo.Echo, addr string, tlsConfig *tls.Config) error {
	lc := &net.ListenConfig{}
	ln, err := lc.Listen(context.Background(), "tcp", addr)
	if err != nil {
		return err
	}
	e.TLSListener = tls.NewListener(ln, tlsConfig)
	e.TLSServer.TLSConfig = tlsConfig
	return e.Server.Serve(e.TLSListener)
}
