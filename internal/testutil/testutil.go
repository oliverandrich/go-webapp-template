// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

// Package testutil provides test helpers and fixtures.
package testutil

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"codeberg.org/oliverandrich/go-webapp-template/internal/database"
	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"codeberg.org/oliverandrich/go-webapp-template/internal/repository"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/vinovest/sqlx"
)

// NewTestDB creates an in-memory SQLite database for tests.
// Returns both the database connection and the repository for convenience.
func NewTestDB(t *testing.T) (*sqlx.DB, *repository.Repository) {
	t.Helper()
	db, err := database.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})
	repo := repository.New(db)
	return db, repo
}

// NewTestUser creates a test user in the database.
func NewTestUser(t *testing.T, repo *repository.Repository, username string) *models.User {
	t.Helper()
	ctx := context.Background()
	user, err := repo.CreateUser(ctx, username)
	require.NoError(t, err)
	return user
}

// NewTestCredential creates a test credential for a user.
func NewTestCredential(t *testing.T, repo *repository.Repository, userID int64, name string) *models.Credential {
	t.Helper()
	ctx := context.Background()
	cred := &models.Credential{
		UserID:       userID,
		CredentialID: []byte("test-credential-id-" + name),
		PublicKey:    []byte("test-public-key"),
		AAGUID:       []byte("test-aaguid-1234"),
		SignCount:    0,
		Name:         name,
	}
	err := repo.CreateCredential(ctx, cred)
	require.NoError(t, err)
	return cred
}

// NewEchoContext creates an Echo context for handler tests.
func NewEchoContext(e *echo.Echo, method, path string, body io.Reader) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// NewEchoContextWithHeaders creates an Echo context with custom headers.
func NewEchoContextWithHeaders(e *echo.Echo, method, path string, body io.Reader, headers map[string]string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, body)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// NewRequest creates an HTTP request for testing.
func NewRequest(method, path string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return req
}
