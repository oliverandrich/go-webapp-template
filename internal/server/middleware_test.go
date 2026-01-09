// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"codeberg.org/oliverandrich/go-webapp-template/internal/appcontext"
	"codeberg.org/oliverandrich/go-webapp-template/internal/config"
	"codeberg.org/oliverandrich/go-webapp-template/internal/i18n"
	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"codeberg.org/oliverandrich/go-webapp-template/internal/repository"
	"codeberg.org/oliverandrich/go-webapp-template/internal/services/session"
	"codeberg.org/oliverandrich/go-webapp-template/internal/testutil"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsHashedAsset(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/static/js/htmx.abc12345.js", true},
		{"/static/css/styles.d073ff63.css", true},
		{"/static/js/htmx.dev.js", false},
		{"/static/js/htmx.js", false},
		{"/static/js/htmx.ABCDEFGH.js", false},  // uppercase not allowed
		{"/static/js/htmx.abcd123.js", false},   // wrong length
		{"/static/js/htmx.abcd12345.js", false}, // wrong length
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, isHashedAsset(tt.path))
		})
	}
}

func TestStaticCacheHeaders(t *testing.T) {
	e := echo.New()
	e.Use(staticCacheHeaders())
	e.GET("/static/*", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	t.Run("hashed asset gets immutable cache", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/static/js/htmx.abc12345.js", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, "public, max-age=31536000, immutable", rec.Header().Get("Cache-Control"))
	})

	t.Run("dev asset gets no-cache", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/static/js/htmx.dev.js", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, "no-cache, no-store, must-revalidate", rec.Header().Get("Cache-Control"))
	})

	t.Run("regular asset gets no cache header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/static/js/htmx.js", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Empty(t, rec.Header().Get("Cache-Control"))
	})
}

func TestI18nMiddleware(t *testing.T) {
	// Initialize i18n bundle
	require.NoError(t, i18n.Init())

	e := echo.New()
	e.Use(i18nMiddleware())

	var locale string
	e.GET("/", func(c echo.Context) error {
		locale = i18n.GetLocale(c.Request().Context())
		return c.NoContent(http.StatusOK)
	})

	t.Run("English header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Language", "en-US")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.True(t, strings.HasPrefix(locale, "en"), "expected locale to start with 'en', got %s", locale)
	})

	t.Run("German header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Language", "de-DE")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.True(t, strings.HasPrefix(locale, "de"), "expected locale to start with 'de', got %s", locale)
	})
}

func TestAuthMiddleware_NoSession(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	sessMgr, err := session.NewManager(&config.SessionConfig{
		CookieName: "_session",
		MaxAge:     3600,
		HashKey:    "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}, false)
	require.NoError(t, err)

	e := echo.New()
	// Create custom context middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &appcontext.Context{Context: c}
			return next(cc)
		}
	})
	e.Use(AuthMiddleware(sessMgr, repo))

	var contextUser *models.User
	e.GET("/", func(c echo.Context) error {
		if cc, ok := c.(*appcontext.Context); ok {
			contextUser = cc.User
		}
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Nil(t, contextUser)
}

func TestAuthMiddleware_WithSession(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	user := testutil.NewTestUser(t, db, "testuser", "Test User")

	sessMgr, err := session.NewManager(&config.SessionConfig{
		CookieName: "_session",
		MaxAge:     3600,
		HashKey:    "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}, false)
	require.NoError(t, err)

	// Create session cookie
	cookie, err := sessMgr.Create(user.ID, user.Username)
	require.NoError(t, err)

	e := echo.New()
	// Create custom context middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &appcontext.Context{Context: c}
			return next(cc)
		}
	})
	e.Use(AuthMiddleware(sessMgr, repo))

	var contextUser *models.User
	e.GET("/", func(c echo.Context) error {
		if cc, ok := c.(*appcontext.Context); ok {
			contextUser = cc.User
		}
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, contextUser)
	assert.Equal(t, user.ID, contextUser.ID)
}

func TestRequireAuth_NotAuthenticated(t *testing.T) {
	e := echo.New()
	// Create custom context middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &appcontext.Context{Context: c}
			return next(cc)
		}
	})
	e.Use(RequireAuth())

	e.GET("/protected", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/auth/login", rec.Header().Get("Location"))
}

func TestRequireAuth_Authenticated(t *testing.T) {
	e := echo.New()
	// Create custom context middleware with user
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &appcontext.Context{
				Context: c,
				User:    &models.User{ID: 1, Username: "test"},
			}
			return next(cc)
		}
	})
	e.Use(RequireAuth())

	e.GET("/protected", func(c echo.Context) error {
		return c.String(http.StatusOK, "protected content")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "protected content", rec.Body.String())
}

func TestCsrfMiddleware(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			BaseURL: "http://localhost:8080",
		},
	}

	mw := csrfMiddleware(cfg)

	assert.NotNil(t, mw)
}

func TestCsrfMiddleware_HTTPS(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			BaseURL: "https://example.com",
		},
	}

	mw := csrfMiddleware(cfg)

	assert.NotNil(t, mw)
}

func TestCsrfToContext(t *testing.T) {
	e := echo.New()
	mw := csrfToContext()
	e.Use(mw)

	var csrfToken string
	e.GET("/", func(c echo.Context) error {
		// Try to get CSRF from context
		if token := c.Request().Context().Value(appcontext.CSRFToken{}); token != nil {
			csrfToken = token.(string)
		}
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// CSRF token is not set without the CSRF middleware, but the middleware should still work
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, csrfToken) // No token without CSRF middleware
}

func TestCsrfToContext_WithToken(t *testing.T) {
	e := echo.New()

	// Middleware that sets a fake CSRF token
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("csrf", "test-token")
			return next(c)
		}
	})
	e.Use(csrfToContext())

	var csrfToken string
	e.GET("/", func(c echo.Context) error {
		if token := c.Request().Context().Value(appcontext.CSRFToken{}); token != nil {
			csrfToken = token.(string)
		}
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test-token", csrfToken)
}

func TestAuthMiddleware_InvalidSession(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	sessMgr, err := session.NewManager(&config.SessionConfig{
		CookieName: "_session",
		MaxAge:     3600,
		HashKey:    "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}, false)
	require.NoError(t, err)

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &appcontext.Context{Context: c}
			return next(cc)
		}
	})
	e.Use(AuthMiddleware(sessMgr, repo))

	var contextUser *models.User
	e.GET("/", func(c echo.Context) error {
		if cc, ok := c.(*appcontext.Context); ok {
			contextUser = cc.User
		}
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Add an invalid cookie
	req.AddCookie(&http.Cookie{
		Name:  "_session",
		Value: "invalid-cookie-data",
	})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Nil(t, contextUser) // No user since cookie was invalid
}

func TestAuthMiddleware_UserNotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	sessMgr, err := session.NewManager(&config.SessionConfig{
		CookieName: "_session",
		MaxAge:     3600,
		HashKey:    "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}, false)
	require.NoError(t, err)

	// Create a valid session cookie for a non-existent user
	cookie, err := sessMgr.Create(99999, "nonexistent")
	require.NoError(t, err)

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &appcontext.Context{Context: c}
			return next(cc)
		}
	})
	e.Use(AuthMiddleware(sessMgr, repo))

	var contextUser *models.User
	e.GET("/", func(c echo.Context) error {
		if cc, ok := c.(*appcontext.Context); ok {
			contextUser = cc.User
		}
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Nil(t, contextUser) // No user since user not in database
}

func TestAuthMiddleware_NotCustomContext(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	sessMgr, err := session.NewManager(&config.SessionConfig{
		CookieName: "_session",
		MaxAge:     3600,
		HashKey:    "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}, false)
	require.NoError(t, err)

	e := echo.New()
	// Don't add custom context middleware - use standard echo.Context
	e.Use(AuthMiddleware(sessMgr, repo))

	e.GET("/", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code) // Should pass through without error
}

func TestRequireAuth_NotCustomContext(t *testing.T) {
	e := echo.New()
	// Don't add custom context middleware - use standard echo.Context
	e.Use(RequireAuth())

	e.GET("/protected", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Should redirect because context is not *Context type
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/auth/login", rec.Header().Get("Location"))
}

func TestStaticCacheHeaders_NonStaticPath(t *testing.T) {
	e := echo.New()
	e.Use(staticCacheHeaders())
	e.GET("/api/data", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Non-static paths should not have cache headers
	assert.Empty(t, rec.Header().Get("Cache-Control"))
}

func TestCustomContext_SetsAssets(t *testing.T) {
	e := echo.New()
	assets := &appcontext.Assets{
		CSSPath: "/static/css/styles.abc123.css",
		JSPath:  "/static/js/htmx.def456.js",
	}

	var capturedContext *appcontext.Context
	handler := func(c echo.Context) error {
		cc, ok := c.(*appcontext.Context)
		require.True(t, ok, "context should be *appcontext.Context")
		capturedContext = cc
		return nil
	}

	middleware := customContext(assets)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := wrappedHandler(c)

	require.NoError(t, err)
	assert.NotNil(t, capturedContext)
	assert.Equal(t, assets, capturedContext.Assets)
	assert.Equal(t, "/static/css/styles.abc123.css", capturedContext.Assets.CSSPath)
	assert.Equal(t, "/static/js/htmx.def456.js", capturedContext.Assets.JSPath)
}

func TestCustomContext_SetsContextValues(t *testing.T) {
	e := echo.New()
	assets := &appcontext.Assets{
		CSSPath: "/static/css/test.css",
		JSPath:  "/static/js/test.js",
	}

	var capturedRequest *http.Request
	handler := func(c echo.Context) error {
		capturedRequest = c.Request()
		return nil
	}

	middleware := customContext(assets)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := wrappedHandler(c)

	require.NoError(t, err)

	// Check context values were set for template access
	ctx := capturedRequest.Context()
	cssPath := ctx.Value(appcontext.CSSPath{})
	jsPath := ctx.Value(appcontext.JSPath{})

	assert.Equal(t, "/static/css/test.css", cssPath)
	assert.Equal(t, "/static/js/test.js", jsPath)
}

func TestCustomContext_ParsesHtmxHeaders(t *testing.T) {
	e := echo.New()
	assets := &appcontext.Assets{
		CSSPath: "/static/css/styles.css",
		JSPath:  "/static/js/htmx.js",
	}

	var capturedContext *appcontext.Context
	handler := func(c echo.Context) error {
		cc, ok := c.(*appcontext.Context)
		require.True(t, ok)
		capturedContext = cc
		return nil
	}

	middleware := customContext(assets)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Target", "main-content")
	req.Header.Set("HX-Trigger", "submit-btn")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := wrappedHandler(c)

	require.NoError(t, err)
	assert.NotNil(t, capturedContext.Htmx)
	assert.True(t, capturedContext.Htmx.IsHtmx)
	assert.Equal(t, "main-content", capturedContext.Htmx.Target)
	assert.Equal(t, "submit-btn", capturedContext.Htmx.Trigger)
}

func TestCustomContext_NonHtmxRequest(t *testing.T) {
	e := echo.New()
	assets := &appcontext.Assets{
		CSSPath: "/static/css/styles.css",
		JSPath:  "/static/js/htmx.js",
	}

	var capturedContext *appcontext.Context
	handler := func(c echo.Context) error {
		cc, ok := c.(*appcontext.Context)
		require.True(t, ok)
		capturedContext = cc
		return nil
	}

	middleware := customContext(assets)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No HX-Request header
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := wrappedHandler(c)

	require.NoError(t, err)
	assert.NotNil(t, capturedContext.Htmx)
	assert.False(t, capturedContext.Htmx.IsHtmx)
}

func TestCustomContext_UserInitiallyNil(t *testing.T) {
	e := echo.New()
	assets := &appcontext.Assets{
		CSSPath: "/static/css/styles.css",
		JSPath:  "/static/js/htmx.js",
	}

	var capturedContext *appcontext.Context
	handler := func(c echo.Context) error {
		cc, ok := c.(*appcontext.Context)
		require.True(t, ok)
		capturedContext = cc
		return nil
	}

	middleware := customContext(assets)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := wrappedHandler(c)

	require.NoError(t, err)
	assert.Nil(t, capturedContext.User)
	assert.False(t, capturedContext.IsAuthenticated())
}

func TestCustomContext_PreservesOriginalContext(t *testing.T) {
	e := echo.New()
	assets := &appcontext.Assets{
		CSSPath: "/static/css/styles.css",
		JSPath:  "/static/js/htmx.js",
	}

	var capturedContext *appcontext.Context
	handler := func(c echo.Context) error {
		cc, ok := c.(*appcontext.Context)
		require.True(t, ok)
		capturedContext = cc
		return nil
	}

	middleware := customContext(assets)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := wrappedHandler(c)

	require.NoError(t, err)
	// Original Echo context methods should still work
	assert.Equal(t, "/api/test", capturedContext.Request().URL.Path)
	assert.Equal(t, "application/json", capturedContext.Request().Header.Get("Content-Type"))
}
