// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package main

import (
	"log/slog"
	"net/http"

	"github.com/oliverandrich/go-webapp-template/internal/config"
	"github.com/oliverandrich/go-webapp-template/internal/csrf"
	"github.com/oliverandrich/go-webapp-template/internal/handlers"
	"github.com/oliverandrich/go-webapp-template/internal/middleware"
	"github.com/oliverandrich/go-webapp-template/internal/repository"
	"github.com/oliverandrich/go-webapp-template/internal/services/auth"
	"github.com/oliverandrich/go-webapp-template/internal/sse"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

// routerDeps holds dependencies needed to set up routes
type routerDeps struct {
	cfg            *config.Config
	repo           *repository.Repository
	sessionManager *scs.SessionManager
	authService    *auth.Service
	sseHub         *sse.Hub
	logger         *slog.Logger
}

// setupRoutes configures all HTTP routes on the given router
func setupRoutes(r chi.Router, deps *routerDeps) {
	cfg := deps.cfg

	// Apply global middlewares (order matters)
	r.Use(middleware.RequestLogger(deps.logger))
	r.Use(middleware.StripTrailingSlash)
	r.Use(csrf.Middleware(cfg.Auth.CookieSecure))
	r.Use(middleware.Locale)
	r.Use(deps.sessionManager.LoadAndSave)
	r.Use(middleware.LoadUser(deps.sessionManager, deps.repo))

	// Static files
	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	// Health check - public
	h := handlers.New(deps.repo)
	r.Get("/health", h.Health)

	// Auth routes - public
	authHandler := handlers.NewAuthHandler(deps.authService, deps.sessionManager, &cfg.Auth)

	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", authHandler.LoginPage)
		r.Post("/login", authHandler.Login)
		r.Get("/register", authHandler.RegisterPage)
		r.Post("/register", authHandler.Register)
		r.Post("/logout", authHandler.Logout)
	})

	// SSE handler for real-time updates
	sseHandler := handlers.NewSSEHandler(deps.sseHub, deps.sessionManager)

	// Protected routes - require authentication
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth)

		r.Get("/", h.Index)
		r.Get("/events", sseHandler.Events)
	})

	// 404 handler
	r.NotFound(handlers.NotFound)
}
