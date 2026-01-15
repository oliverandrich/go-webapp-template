// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package templates

import (
	"context"

	"github.com/oliverandrich/go-webapp-template/internal/appcontext"
	"github.com/oliverandrich/go-webapp-template/internal/i18n"
	"github.com/oliverandrich/go-webapp-template/internal/models"
)

// CSRFToken returns the CSRF token from the context.
func CSRFToken(ctx context.Context) string {
	if token, ok := ctx.Value(appcontext.CSRFToken{}).(string); ok {
		return token
	}
	return ""
}

// T translates a message by ID.
func T(ctx context.Context, messageID string) string {
	return i18n.T(ctx, messageID)
}

// TData translates a message with template data.
func TData(ctx context.Context, messageID string, data map[string]any) string {
	return i18n.TData(ctx, messageID, data)
}

// Locale returns the current locale.
func Locale(ctx context.Context) string {
	return i18n.GetLocale(ctx)
}

// CSSPath returns the path to the hashed CSS file.
func CSSPath(ctx context.Context) string {
	if path, ok := ctx.Value(appcontext.CSSPath{}).(string); ok {
		return path
	}
	return "/static/css/styles.css"
}

// JSPath returns the path to the hashed htmx JS file.
func JSPath(ctx context.Context) string {
	if path, ok := ctx.Value(appcontext.JSPath{}).(string); ok {
		return path
	}
	return "/static/js/htmx.js"
}

// GetUser returns the authenticated user from context, or nil if not logged in.
func GetUser(ctx context.Context) *models.User {
	if user, ok := ctx.Value(appcontext.User{}).(*models.User); ok {
		return user
	}
	return nil
}

// IsAuthenticated returns true if a user is logged in.
func IsAuthenticated(ctx context.Context) bool {
	return GetUser(ctx) != nil
}
