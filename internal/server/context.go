// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package server

import (
	"context"

	"codeberg.org/oliverandrich/go-webapp-template/internal/ctxkeys"
	"codeberg.org/oliverandrich/go-webapp-template/internal/htmx"
	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"github.com/labstack/echo/v4"
)

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

// customContext wraps the Echo context with our custom Context.
// It also populates request.Context with asset paths for template access.
func customContext(assets *Assets) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Add asset paths to request context (for templates)
			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, ctxkeys.CSSPath{}, assets.CSSPath)
			ctx = context.WithValue(ctx, ctxkeys.JSPath{}, assets.JSPath)
			c.SetRequest(c.Request().WithContext(ctx))

			// Wrap with custom context (for handlers)
			cc := &Context{
				Context: c,
				Htmx:    htmx.ParseRequest(c.Request()),
				Assets:  assets,
			}
			return next(cc)
		}
	}
}
