// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package webauthn_test

import (
	"sync"
	"testing"

	gowebauthn "github.com/go-webauthn/webauthn/webauthn"
	"github.com/oliverandrich/go-webapp-template/internal/config"
	"github.com/oliverandrich/go-webapp-template/internal/services/webauthn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestConfig() *config.WebAuthnConfig {
	return &config.WebAuthnConfig{
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
		RPDisplayName: "Test App",
	}
}

func TestNewService(t *testing.T) {
	cfg := newTestConfig()

	svc, err := webauthn.NewService(cfg)

	require.NoError(t, err)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.WebAuthn())
}

func TestWebAuthn(t *testing.T) {
	cfg := newTestConfig()
	svc, err := webauthn.NewService(cfg)
	require.NoError(t, err)

	wa := svc.WebAuthn()

	assert.NotNil(t, wa)
}

func TestStoreAndGetRegistrationSession(t *testing.T) {
	cfg := newTestConfig()
	svc, err := webauthn.NewService(cfg)
	require.NoError(t, err)

	sessionData := &gowebauthn.SessionData{
		Challenge: "test-challenge",
	}

	svc.StoreRegistrationSession(123, sessionData)

	retrieved, err := svc.GetRegistrationSession(123)

	require.NoError(t, err)
	assert.Equal(t, "test-challenge", retrieved.Challenge)
}

func TestGetRegistrationSession_NotFound(t *testing.T) {
	cfg := newTestConfig()
	svc, err := webauthn.NewService(cfg)
	require.NoError(t, err)

	_, err = svc.GetRegistrationSession(999)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestGetRegistrationSession_DeletesAfterGet(t *testing.T) {
	cfg := newTestConfig()
	svc, err := webauthn.NewService(cfg)
	require.NoError(t, err)

	sessionData := &gowebauthn.SessionData{
		Challenge: "test-challenge",
	}

	svc.StoreRegistrationSession(123, sessionData)

	// First get should succeed
	_, err = svc.GetRegistrationSession(123)
	require.NoError(t, err)

	// Second get should fail (session was deleted)
	_, err = svc.GetRegistrationSession(123)
	assert.Error(t, err)
}

func TestStoreAndGetLoginSession(t *testing.T) {
	cfg := newTestConfig()
	svc, err := webauthn.NewService(cfg)
	require.NoError(t, err)

	sessionData := &gowebauthn.SessionData{
		Challenge: "login-challenge",
	}

	svc.StoreLoginSession(456, sessionData)

	retrieved, err := svc.GetLoginSession(456)

	require.NoError(t, err)
	assert.Equal(t, "login-challenge", retrieved.Challenge)
}

func TestGetLoginSession_NotFound(t *testing.T) {
	cfg := newTestConfig()
	svc, err := webauthn.NewService(cfg)
	require.NoError(t, err)

	_, err = svc.GetLoginSession(999)

	assert.Error(t, err)
}

func TestStoreAndGetDiscoverableSession(t *testing.T) {
	cfg := newTestConfig()
	svc, err := webauthn.NewService(cfg)
	require.NoError(t, err)

	sessionData := &gowebauthn.SessionData{
		Challenge: "discoverable-challenge",
	}

	svc.StoreDiscoverableSession("session-123", sessionData)

	retrieved, err := svc.GetDiscoverableSession("session-123")

	require.NoError(t, err)
	assert.Equal(t, "discoverable-challenge", retrieved.Challenge)
}

func TestGetDiscoverableSession_NotFound(t *testing.T) {
	cfg := newTestConfig()
	svc, err := webauthn.NewService(cfg)
	require.NoError(t, err)

	_, err = svc.GetDiscoverableSession("nonexistent")

	assert.Error(t, err)
}

func TestConcurrentAccess(t *testing.T) {
	cfg := newTestConfig()
	svc, err := webauthn.NewService(cfg)
	require.NoError(t, err)

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Store and retrieve sessions concurrently
	for i := range 50 {
		wg.Add(2)

		// Store goroutine
		go func(id int64) {
			defer wg.Done()
			svc.StoreRegistrationSession(id, &gowebauthn.SessionData{
				Challenge: "challenge",
			})
		}(int64(i))

		// Store a different type
		go func(id int64) {
			defer wg.Done()
			svc.StoreLoginSession(id+1000, &gowebauthn.SessionData{
				Challenge: "login-challenge",
			})
		}(int64(i))
	}

	wg.Wait()

	// Now retrieve them
	for i := range 50 {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			_, err := svc.GetRegistrationSession(id)
			if err != nil {
				errors <- err
			}
		}(int64(i))
	}

	wg.Wait()
	close(errors)

	// Should have no errors
	errorCount := 0
	for range errors {
		errorCount++
	}
	assert.Zero(t, errorCount)
}

func TestSessionIsolation(t *testing.T) {
	cfg := newTestConfig()
	svc, err := webauthn.NewService(cfg)
	require.NoError(t, err)

	// Store sessions of different types with same user ID
	svc.StoreRegistrationSession(123, &gowebauthn.SessionData{Challenge: "reg"})
	svc.StoreLoginSession(123, &gowebauthn.SessionData{Challenge: "login"})
	svc.StoreDiscoverableSession("123", &gowebauthn.SessionData{Challenge: "discover"})

	// Each should be retrievable independently
	reg, err := svc.GetRegistrationSession(123)
	require.NoError(t, err)
	assert.Equal(t, "reg", reg.Challenge)

	login, err := svc.GetLoginSession(123)
	require.NoError(t, err)
	assert.Equal(t, "login", login.Challenge)

	discover, err := svc.GetDiscoverableSession("123")
	require.NoError(t, err)
	assert.Equal(t, "discover", discover.Challenge)
}

func TestOverwriteSession(t *testing.T) {
	cfg := newTestConfig()
	svc, err := webauthn.NewService(cfg)
	require.NoError(t, err)

	// Store initial session
	svc.StoreRegistrationSession(123, &gowebauthn.SessionData{Challenge: "first"})

	// Overwrite with new session
	svc.StoreRegistrationSession(123, &gowebauthn.SessionData{Challenge: "second"})

	// Should get the second one
	retrieved, err := svc.GetRegistrationSession(123)
	require.NoError(t, err)
	assert.Equal(t, "second", retrieved.Challenge)
}
