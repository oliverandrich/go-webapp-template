// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package config

import (
	"fmt"

	"github.com/urfave/cli/v3"
)

// RegistrationMode defines how user registration works
type RegistrationMode string

const (
	// RegistrationInternal allows registration without email verification (LAN/family use)
	RegistrationInternal RegistrationMode = "internal"
	// RegistrationOpen allows open registration (public internet)
	RegistrationOpen RegistrationMode = "open"
	// RegistrationClosed disables self-registration, only admin can create users
	RegistrationClosed RegistrationMode = "closed"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Log      LogConfig
}

type LogConfig struct {
	Level  string // debug, info, warn, error
	Format string // text, json
}

type ServerConfig struct {
	Host    string
	Port    int
	BaseURL string
}

type DatabaseConfig struct {
	DSN string // SQLite DSN (file path or :memory:)
}

type AuthConfig struct {
	RegistrationMode   RegistrationMode
	SessionSecret      string
	SessionDuration    int    // Hours for normal sessions (default: 24)
	RememberMeDuration int    // Hours for "remember me" sessions (default: 720 = 30 days)
	CookieName         string // Session cookie name (default: "session")
	CookieSecure       bool   // HTTPS only cookie
}

// NewFromCLI creates a Config from urfave/cli command flags
func NewFromCLI(cmd *cli.Command) *Config {
	cfg := &Config{
		Server: ServerConfig{
			Host:    cmd.String("host"),
			Port:    int(cmd.Int("port")),
			BaseURL: cmd.String("base-url"),
		},
		Auth: AuthConfig{
			RegistrationMode:   RegistrationMode(cmd.String("registration-mode")),
			SessionSecret:      cmd.String("session-secret"),
			SessionDuration:    int(cmd.Int("session-duration")),
			RememberMeDuration: int(cmd.Int("remember-me-duration")),
			CookieName:         cmd.String("cookie-name"),
			CookieSecure:       cmd.Bool("cookie-secure"),
		},
		Log: LogConfig{
			Level:  cmd.String("log-level"),
			Format: cmd.String("log-format"),
		},
	}

	// BaseURL: auto-generate if not set
	if cfg.Server.BaseURL == "" {
		cfg.Server.BaseURL = fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
	}

	// Database path
	cfg.Database.DSN = cmd.String("database")

	return cfg
}

// IsRegistrationEnabled returns true if users can self-register
func (c *AuthConfig) IsRegistrationEnabled() bool {
	return c.RegistrationMode != RegistrationClosed
}
