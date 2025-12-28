// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package csrf

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/justinas/nosurf"
)

type csrfContextKey struct{}

// ContextKey is the context key for CSRF tokens
var ContextKey = csrfContextKey{}

// GetToken retrieves the CSRF token from the context
func GetToken(ctx context.Context) string {
	if token, ok := ctx.Value(ContextKey).(string); ok {
		return token
	}
	return ""
}

// Middleware wraps nosurf and adds the CSRF token to the context for templates.
func Middleware(secure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		csrfHandler := nosurf.New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add CSRF token to context for templates
			token := nosurf.Token(r)
			ctx := context.WithValue(r.Context(), ContextKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		}))

		// Configure the CSRF cookie
		csrfHandler.SetBaseCookie(http.Cookie{
			HttpOnly: true,
			Path:     "/",
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		})

		// Set custom failure handler
		csrfHandler.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Warn("csrf_failure",
				"path", r.URL.Path,
				"method", r.Method,
				"ip", getClientIP(r),
			)
			http.Error(w, "Invalid CSRF token. Please reload the page.", http.StatusForbidden)
		}))

		return csrfHandler
	}
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}
