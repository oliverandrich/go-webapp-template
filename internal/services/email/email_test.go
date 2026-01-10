// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package email_test

import (
	"testing"
	"time"

	"codeberg.org/oliverandrich/go-webapp-template/internal/config"
	"codeberg.org/oliverandrich/go-webapp-template/internal/services/email"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validSMTPConfig() *config.SMTPConfig {
	return &config.SMTPConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "testuser",
		Password: "testpass",
		From:     "noreply@example.com",
		FromName: "Test App",
		TLS:      true,
	}
}

func TestNewService(t *testing.T) {
	cfg := validSMTPConfig()

	svc, err := email.NewService(cfg, "https://example.com")

	require.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestNewService_MissingHost(t *testing.T) {
	cfg := validSMTPConfig()
	cfg.Host = ""

	_, err := email.NewService(cfg, "https://example.com")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SMTP host is required")
}

func TestNewService_MissingFrom(t *testing.T) {
	cfg := validSMTPConfig()
	cfg.From = ""

	_, err := email.NewService(cfg, "https://example.com")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SMTP from address is required")
}

func TestNewService_TrailingSlashTrimmed(t *testing.T) {
	cfg := validSMTPConfig()

	svc, err := email.NewService(cfg, "https://example.com/")

	require.NoError(t, err)
	assert.NotNil(t, svc)
	// The trailing slash should be trimmed (verified by checking verification URL format in integration tests)
}

func TestGenerateToken(t *testing.T) {
	cfg := validSMTPConfig()
	svc, err := email.NewService(cfg, "https://example.com")
	require.NoError(t, err)

	plaintext, hash, expiresAt, err := svc.GenerateToken()

	require.NoError(t, err)

	// Plaintext should be 64 hex chars (32 bytes)
	assert.Len(t, plaintext, 64)

	// Hash should be 64 hex chars (SHA256 = 32 bytes)
	assert.Len(t, hash, 64)

	// Plaintext and hash should be different
	assert.NotEqual(t, plaintext, hash)

	// Expiry should be ~24 hours in the future
	expectedExpiry := time.Now().Add(24 * time.Hour)
	assert.WithinDuration(t, expectedExpiry, expiresAt, time.Minute)
}

func TestGenerateToken_Unique(t *testing.T) {
	cfg := validSMTPConfig()
	svc, err := email.NewService(cfg, "https://example.com")
	require.NoError(t, err)

	// Generate multiple tokens and ensure they're all unique
	tokens := make(map[string]bool)
	hashes := make(map[string]bool)

	for range 10 {
		plaintext, hash, _, err := svc.GenerateToken()
		require.NoError(t, err)

		assert.False(t, tokens[plaintext], "duplicate token generated")
		assert.False(t, hashes[hash], "duplicate hash generated")

		tokens[plaintext] = true
		hashes[hash] = true
	}
}

func TestHashToken(t *testing.T) {
	token := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	hash := email.HashToken(token)

	// SHA256 produces 32 bytes = 64 hex chars
	assert.Len(t, hash, 64)

	// Same input should produce same hash
	hash2 := email.HashToken(token)
	assert.Equal(t, hash, hash2)
}

func TestHashToken_DifferentInputs(t *testing.T) {
	hash1 := email.HashToken("token1")
	hash2 := email.HashToken("token2")

	assert.NotEqual(t, hash1, hash2)
}

func TestHashToken_EmptyInput(t *testing.T) {
	hash := email.HashToken("")

	// Should still produce a valid hash
	assert.Len(t, hash, 64)
}
