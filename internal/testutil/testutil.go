// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

// Package testutil provides test helpers and fixtures.
package testutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewTestDB creates an in-memory SQLite database for tests.
func NewTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(models.AllModels()...))
	return db
}

// NewTestUser creates a test user in the database.
func NewTestUser(t *testing.T, db *gorm.DB, username string) *models.User {
	t.Helper()
	user := &models.User{
		Username: username,
	}
	require.NoError(t, db.Create(user).Error)
	return user
}

// NewTestCredential creates a test credential for a user.
func NewTestCredential(t *testing.T, db *gorm.DB, userID int64, name string) *models.Credential {
	t.Helper()
	cred := &models.Credential{
		UserID:       userID,
		CredentialID: []byte("test-credential-id-" + name),
		PublicKey:    []byte("test-public-key"),
		AAGUID:       []byte("test-aaguid-1234"),
		SignCount:    0,
		Name:         name,
	}
	require.NoError(t, db.Create(cred).Error)
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
