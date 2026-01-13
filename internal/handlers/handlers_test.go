// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"codeberg.org/oliverandrich/go-webapp-template/internal/handlers"
	"codeberg.org/oliverandrich/go-webapp-template/internal/i18n"
	"codeberg.org/oliverandrich/go-webapp-template/internal/testutil"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func init() {
	// Initialize i18n for template rendering
	_ = i18n.Init()
}

func TestNew(t *testing.T) {
	_, repo := testutil.NewTestDB(t)

	h := handlers.New(repo)

	assert.NotNil(t, h)
}

func TestHealth(t *testing.T) {
	h := handlers.New(nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Health(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
}

func TestHome(t *testing.T) {
	_, repo := testutil.NewTestDB(t)
	h := handlers.New(repo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Add i18n context
	ctx := i18n.WithLocale(req.Context(), language.English)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Home(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<!doctype html>")
}

func TestDashboard(t *testing.T) {
	_, repo := testutil.NewTestDB(t)
	h := handlers.New(repo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	// Add i18n context
	ctx := i18n.WithLocale(req.Context(), language.English)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Dashboard(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<!doctype html>")
}
