// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"codeberg.org/oliverandrich/go-webapp-template/internal/config"
	"codeberg.org/oliverandrich/go-webapp-template/internal/ctxkeys"
	"codeberg.org/oliverandrich/go-webapp-template/internal/i18n"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func setupMiddleware(e *echo.Echo, cfg *config.Config, assets *Assets) {
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(requestLogger())
	e.Use(middleware.Secure())
	e.Use(middleware.Gzip())
	e.Use(middleware.BodyLimit(fmt.Sprintf("%dM", cfg.Server.MaxBodySize)))
	e.Use(staticCacheHeaders())
	e.Use(csrfMiddleware(cfg))
	e.Use(csrfToContext())
	e.Use(assetsToContext(assets))
	e.Use(i18nMiddleware())
}

// csrfMiddleware configures CSRF protection.
func csrfMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	secure := strings.HasPrefix(cfg.Server.BaseURL, "https://")

	return middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "form:csrf_token,header:X-CSRF-Token",
		CookieName:     "_csrf",
		CookiePath:     "/",
		CookieSecure:   secure,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteLaxMode,
	})
}

// csrfToContext copies the CSRF token to the request context.
func csrfToContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if token, ok := c.Get("csrf").(string); ok {
				ctx := context.WithValue(c.Request().Context(), ctxkeys.CSRFToken{}, token)
				c.SetRequest(c.Request().WithContext(ctx))
			}
			return next(c)
		}
	}
}

// requestLogger returns middleware that logs requests using slog.
func requestLogger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogMethod:   true,
		LogLatency:  true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			attrs := []slog.Attr{
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.Duration("latency", v.Latency),
			}

			if v.Error != nil {
				attrs = append(attrs, slog.String("error", v.Error.Error()))
				slog.LogAttrs(c.Request().Context(), slog.LevelError, "request", attrs...)
			} else {
				slog.LogAttrs(c.Request().Context(), slog.LevelInfo, "request", attrs...)
			}

			return nil
		},
	})
}

// i18nMiddleware sets the locale based on Accept-Language header.
func i18nMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			acceptLang := c.Request().Header.Get("Accept-Language")
			lang := i18n.MatchLanguage(acceptLang)
			ctx := i18n.WithLocale(c.Request().Context(), lang)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

// staticCacheHeaders adds cache headers for static assets.
func staticCacheHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if strings.HasPrefix(path, "/static/") {
				if isHashedAsset(path) {
					// Hashed assets get immutable caching
					c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				} else if strings.Contains(path, ".dev.") {
					// Dev assets never cache
					c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				}
			}
			return next(c)
		}
	}
}

// isHashedAsset checks if the path contains a hash pattern like .abc12345.
func isHashedAsset(path string) bool {
	// Match pattern: name.HASH.ext where HASH is 8 hex characters
	parts := strings.Split(path, ".")
	if len(parts) >= 3 {
		hash := parts[len(parts)-2]
		if len(hash) == 8 {
			for _, c := range hash {
				isDigit := c >= '0' && c <= '9'
				isHexLetter := c >= 'a' && c <= 'f'
				if !isDigit && !isHexLetter {
					return false
				}
			}
			return true
		}
	}
	return false
}
