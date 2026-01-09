// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package templates

import (
	"context"

	"codeberg.org/oliverandrich/go-webapp-template/internal/ctxkeys"
	"codeberg.org/oliverandrich/go-webapp-template/internal/i18n"
)

// CSRFToken returns the CSRF token from the context.
func CSRFToken(ctx context.Context) string {
	if token, ok := ctx.Value(ctxkeys.CSRFToken{}).(string); ok {
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
	if path, ok := ctx.Value(ctxkeys.CSSPath{}).(string); ok {
		return path
	}
	return "/static/css/styles.css"
}

// JSPath returns the path to the hashed htmx JS file.
func JSPath(ctx context.Context) string {
	if path, ok := ctx.Value(ctxkeys.JSPath{}).(string); ok {
		return path
	}
	return "/static/js/htmx.js"
}
