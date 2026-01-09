// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package handlers

import (
	"net/http"

	"codeberg.org/oliverandrich/go-webapp-template/internal/repository"
	"codeberg.org/oliverandrich/go-webapp-template/internal/templates"
	"github.com/labstack/echo/v4"
)

// Handlers contains all HTTP handlers.
type Handlers struct {
	repo *repository.Repository
}

// New creates a new Handlers instance.
func New(repo *repository.Repository) *Handlers {
	return &Handlers{repo: repo}
}

// Health returns the health status.
func (h *Handlers) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// Home renders the home page.
func (h *Handlers) Home(c echo.Context) error {
	return Render(c, http.StatusOK, templates.Home())
}
