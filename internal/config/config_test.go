// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestIsLocalhost(t *testing.T) {
	tests := []struct {
		host     string
		expected bool
	}{
		{"", true},
		{"localhost", true},
		{"127.0.0.1", true},
		{"::1", true},
		{"app.localhost", true},
		{"sub.domain.localhost", true},
		{"example.com", false},
		{"www.example.com", false},
		{"192.168.1.1", false},
		{"localhost.com", false}, // not a real localhost
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsLocalhost(tt.host))
		})
	}
}

func TestShouldUseTLS(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		host     string
		expected bool
	}{
		{"off mode", "off", "example.com", false},
		{"acme mode", "acme", "localhost", true},
		{"selfsigned mode", "selfsigned", "localhost", true},
		{"manual mode", "manual", "localhost", true},
		{"auto mode with localhost", "auto", "localhost", false},
		{"auto mode with remote host", "auto", "example.com", true},
		{"empty mode with localhost", "", "localhost", false},
		{"empty mode with remote host", "", "example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, shouldUseTLS(tt.mode, tt.host))
		})
	}
}

func TestBuildBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		expected string
	}{
		{
			name: "localhost HTTP default port",
			cfg: &Config{
				Server: ServerConfig{Host: "localhost", Port: 80},
				TLS:    TLSConfig{Mode: "off"},
			},
			expected: "http://localhost",
		},
		{
			name: "localhost HTTP custom port",
			cfg: &Config{
				Server: ServerConfig{Host: "localhost", Port: 8080},
				TLS:    TLSConfig{Mode: "off"},
			},
			expected: "http://localhost:8080",
		},
		{
			name: "remote host with auto TLS",
			cfg: &Config{
				Server: ServerConfig{Host: "example.com", Port: 443},
				TLS:    TLSConfig{Mode: "auto"},
			},
			expected: "https://example.com",
		},
		{
			name: "remote host with auto TLS custom port",
			cfg: &Config{
				Server: ServerConfig{Host: "example.com", Port: 8443},
				TLS:    TLSConfig{Mode: "selfsigned"},
			},
			expected: "https://example.com:8443",
		},
		{
			name: "ACME mode forces port 443",
			cfg: &Config{
				Server: ServerConfig{Host: "example.com", Port: 8080},
				TLS:    TLSConfig{Mode: "acme"},
			},
			expected: "https://example.com",
		},
		{
			name: "localhost with auto TLS uses HTTP",
			cfg: &Config{
				Server: ServerConfig{Host: "localhost", Port: 8080},
				TLS:    TLSConfig{Mode: "auto"},
			},
			expected: "http://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, buildBaseURL(tt.cfg))
		})
	}
}

func TestApplyWebAuthnDefaults(t *testing.T) {
	t.Run("applies defaults when empty", func(t *testing.T) {
		cfg := &Config{
			Server: ServerConfig{
				Host:    "localhost",
				BaseURL: "http://localhost:8080",
			},
			WebAuthn: WebAuthnConfig{},
		}

		applyWebAuthnDefaults(cfg)

		assert.Equal(t, "localhost", cfg.WebAuthn.RPID)
		assert.Equal(t, "http://localhost:8080", cfg.WebAuthn.RPOrigin)
		assert.Equal(t, "Go Web App", cfg.WebAuthn.RPDisplayName)
	})

	t.Run("does not override existing values", func(t *testing.T) {
		cfg := &Config{
			Server: ServerConfig{
				Host:    "localhost",
				BaseURL: "http://localhost:8080",
			},
			WebAuthn: WebAuthnConfig{
				RPID:          "custom.domain",
				RPOrigin:      "https://custom.domain",
				RPDisplayName: "Custom App",
			},
		}

		applyWebAuthnDefaults(cfg)

		assert.Equal(t, "custom.domain", cfg.WebAuthn.RPID)
		assert.Equal(t, "https://custom.domain", cfg.WebAuthn.RPOrigin)
		assert.Equal(t, "Custom App", cfg.WebAuthn.RPDisplayName)
	})
}

func TestFlags(t *testing.T) {
	flags := Flags()

	// Should have all expected flags
	assert.NotEmpty(t, flags)

	// Check for key flags
	flagNames := make(map[string]bool)
	for _, f := range flags {
		for _, name := range f.Names() {
			flagNames[name] = true
		}
	}

	assert.True(t, flagNames["host"], "should have host flag")
	assert.True(t, flagNames["port"], "should have port flag")
	assert.True(t, flagNames["base-url"], "should have base-url flag")
	assert.True(t, flagNames["log-level"], "should have log-level flag")
	assert.True(t, flagNames["database-dsn"], "should have database-dsn flag")
	assert.True(t, flagNames["tls-mode"], "should have tls-mode flag")
	assert.True(t, flagNames["webauthn-rp-id"], "should have webauthn-rp-id flag")
	assert.True(t, flagNames["session-cookie-name"], "should have session-cookie-name flag")
}

func TestNewFromCLI(t *testing.T) {
	app := &cli.Command{
		Name:  "test",
		Flags: Flags(),
		Action: func(_ context.Context, cmd *cli.Command) error {
			cfg := NewFromCLI(cmd)

			// Verify defaults are applied
			assert.NotNil(t, cfg)
			assert.Equal(t, "localhost", cfg.Server.Host)
			assert.Equal(t, 8080, cfg.Server.Port)
			assert.Equal(t, "info", cfg.Log.Level)
			assert.Equal(t, "text", cfg.Log.Format)
			assert.Equal(t, "_session", cfg.Session.CookieName)
			assert.Equal(t, 604800, cfg.Session.MaxAge) // 7 days in seconds

			// BaseURL should be auto-generated
			assert.NotEmpty(t, cfg.Server.BaseURL)

			// WebAuthn defaults should be applied
			assert.Equal(t, "localhost", cfg.WebAuthn.RPID)
			assert.Equal(t, "Go Web App", cfg.WebAuthn.RPDisplayName)

			return nil
		},
	}

	// Run the command with default flags
	err := app.Run(context.Background(), []string{"test"})
	assert.NoError(t, err)
}

func TestNewFromCLI_WithCustomValues(t *testing.T) {
	app := &cli.Command{
		Name:  "test",
		Flags: Flags(),
		Action: func(_ context.Context, cmd *cli.Command) error {
			cfg := NewFromCLI(cmd)

			// Verify custom values
			assert.Equal(t, "0.0.0.0", cfg.Server.Host)
			assert.Equal(t, 9000, cfg.Server.Port)
			assert.Equal(t, "https://example.com", cfg.Server.BaseURL)
			assert.Equal(t, "debug", cfg.Log.Level)
			assert.Equal(t, "./data/test.db", cfg.Database.DSN)

			return nil
		},
	}

	// Run with custom args
	args := []string{
		"test",
		"--host", "0.0.0.0",
		"--port", "9000",
		"--base-url", "https://example.com",
		"--log-level", "debug",
		"--database-dsn", "./data/test.db",
	}
	err := app.Run(context.Background(), args)
	assert.NoError(t, err)
}
