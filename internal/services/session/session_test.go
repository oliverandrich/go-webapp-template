// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package session_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"codeberg.org/oliverandrich/go-webapp-template/internal/config"
	"codeberg.org/oliverandrich/go-webapp-template/internal/services/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validHashKey is a valid 32-byte hex-encoded key for testing
const validHashKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

// validBlockKey is a valid 32-byte hex-encoded key for encryption testing
const validBlockKey = "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"

func newTestConfig() *config.SessionConfig {
	return &config.SessionConfig{
		CookieName: "_test_session",
		MaxAge:     3600, // 1 hour
		HashKey:    validHashKey,
	}
}

func TestNewManager(t *testing.T) {
	cfg := newTestConfig()

	mgr, err := session.NewManager(cfg, false)

	require.NoError(t, err)
	assert.NotNil(t, mgr)
}

func TestNewManager_WithBlockKey(t *testing.T) {
	cfg := newTestConfig()
	cfg.BlockKey = validBlockKey

	mgr, err := session.NewManager(cfg, true)

	require.NoError(t, err)
	assert.NotNil(t, mgr)
}

func TestNewManager_InvalidHashKey_NotHex(t *testing.T) {
	cfg := newTestConfig()
	cfg.HashKey = "not-hex-encoded"

	_, err := session.NewManager(cfg, false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid session hash key")
}

func TestNewManager_InvalidHashKey_WrongLength(t *testing.T) {
	cfg := newTestConfig()
	cfg.HashKey = "0123456789abcdef" // only 8 bytes

	_, err := session.NewManager(cfg, false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be 32 bytes")
}

func TestNewManager_InvalidBlockKey_NotHex(t *testing.T) {
	cfg := newTestConfig()
	cfg.BlockKey = "not-hex-encoded"

	_, err := session.NewManager(cfg, false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid session block key")
}

func TestNewManager_InvalidBlockKey_WrongLength(t *testing.T) {
	cfg := newTestConfig()
	cfg.BlockKey = "0123456789abcdef" // only 8 bytes

	_, err := session.NewManager(cfg, false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be 32 bytes")
}

func TestNewManager_DevMode_GeneratesKey(t *testing.T) {
	cfg := &config.SessionConfig{
		CookieName: "_session",
		MaxAge:     3600,
		HashKey:    "", // empty - should auto-generate
	}

	mgr, err := session.NewManager(cfg, false)

	require.NoError(t, err)
	assert.NotNil(t, mgr)
}

func TestCreate(t *testing.T) {
	cfg := newTestConfig()
	mgr, err := session.NewManager(cfg, false)
	require.NoError(t, err)

	cookie, err := mgr.Create(123, "testuser")

	require.NoError(t, err)
	assert.Equal(t, "_test_session", cookie.Name)
	assert.NotEmpty(t, cookie.Value)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, 3600, cookie.MaxAge)
	assert.True(t, cookie.HttpOnly)
	assert.False(t, cookie.Secure)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
}

func TestCreate_SecureMode(t *testing.T) {
	cfg := newTestConfig()
	mgr, err := session.NewManager(cfg, true)
	require.NoError(t, err)

	cookie, err := mgr.Create(123, "testuser")

	require.NoError(t, err)
	assert.True(t, cookie.Secure)
}

func TestParse(t *testing.T) {
	cfg := newTestConfig()
	mgr, err := session.NewManager(cfg, false)
	require.NoError(t, err)

	// Create a session
	cookie, err := mgr.Create(123, "testuser")
	require.NoError(t, err)

	// Create a request with the cookie
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)

	// Parse the session
	data, err := mgr.Parse(req)

	require.NoError(t, err)
	require.NotNil(t, data)
	assert.Equal(t, int64(123), data.UserID)
	assert.Equal(t, "testuser", data.Username)
	assert.False(t, data.ExpiresAt.IsZero())
}

func TestParse_NoCookie(t *testing.T) {
	cfg := newTestConfig()
	mgr, err := session.NewManager(cfg, false)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	data, err := mgr.Parse(req)

	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestParse_InvalidCookie(t *testing.T) {
	cfg := newTestConfig()
	mgr, err := session.NewManager(cfg, false)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "_test_session",
		Value: "invalid-cookie-value",
	})

	data, err := mgr.Parse(req)

	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestParse_TamperedCookie(t *testing.T) {
	cfg := newTestConfig()
	mgr, err := session.NewManager(cfg, false)
	require.NoError(t, err)

	// Create a valid session
	cookie, err := mgr.Create(123, "testuser")
	require.NoError(t, err)

	// Tamper with the cookie value
	cookie.Value = cookie.Value[:len(cookie.Value)-5] + "XXXXX"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)

	data, err := mgr.Parse(req)

	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestParse_ExpiredSession(t *testing.T) {
	cfg := &config.SessionConfig{
		CookieName: "_test_session",
		MaxAge:     1, // 1 second
		HashKey:    validHashKey,
	}
	mgr, err := session.NewManager(cfg, false)
	require.NoError(t, err)

	// Create a session
	cookie, err := mgr.Create(123, "testuser")
	require.NoError(t, err)

	// Wait for expiration
	time.Sleep(2 * time.Second)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)

	data, err := mgr.Parse(req)

	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestParse_DifferentManager(t *testing.T) {
	cfg1 := newTestConfig()
	mgr1, err := session.NewManager(cfg1, false)
	require.NoError(t, err)

	// Create a session with manager 1
	cookie, err := mgr1.Create(123, "testuser")
	require.NoError(t, err)

	// Create a different manager with different key
	cfg2 := &config.SessionConfig{
		CookieName: "_test_session",
		MaxAge:     3600,
		HashKey:    validBlockKey, // different key
	}
	mgr2, err := session.NewManager(cfg2, false)
	require.NoError(t, err)

	// Try to parse with manager 2
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)

	data, err := mgr2.Parse(req)

	require.NoError(t, err)
	assert.Nil(t, data) // Should not be able to decode
}

func TestClear(t *testing.T) {
	cfg := newTestConfig()
	mgr, err := session.NewManager(cfg, false)
	require.NoError(t, err)

	cookie := mgr.Clear()

	assert.Equal(t, "_test_session", cookie.Name)
	assert.Empty(t, cookie.Value)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, -1, cookie.MaxAge)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
}

func TestClear_SecureMode(t *testing.T) {
	cfg := newTestConfig()
	mgr, err := session.NewManager(cfg, true)
	require.NoError(t, err)

	cookie := mgr.Clear()

	assert.True(t, cookie.Secure)
}
