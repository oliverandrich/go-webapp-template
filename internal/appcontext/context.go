// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

// Package appcontext provides the custom Echo context and context keys.
package appcontext

import (
	"codeberg.org/oliverandrich/go-webapp-template/internal/htmx"
	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"github.com/labstack/echo/v4"
)

// Context keys for storing values in context.Context.
type (
	// CSRFToken is the context key for the CSRF token.
	CSRFToken struct{}
	// CSSPath is the context key for the CSS path.
	CSSPath struct{}
	// JSPath is the context key for the JS (htmx) path.
	JSPath struct{}
	// User is the context key for the authenticated user.
	User struct{}
)

// Assets holds paths to static assets.
type Assets struct {
	CSSPath string
	JSPath  string
}

// Context is a custom Echo context with typed fields for htmx, assets, and user.
type Context struct {
	echo.Context
	Htmx   *htmx.Request
	Assets *Assets
	User   *models.User // nil if not authenticated
}

// GetUser returns the authenticated user, or nil if not authenticated.
func (c *Context) GetUser() *models.User {
	return c.User
}

// IsAuthenticated returns true if the user is authenticated.
func (c *Context) IsAuthenticated() bool {
	return c.User != nil
}
