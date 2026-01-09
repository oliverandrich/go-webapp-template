// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package config

import (
	"fmt"
	"strings"

	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
)

var configFile = altsrc.StringSourcer("config.toml")

type Config struct { //nolint:govet // fieldalignment not critical for config structs
	Server   ServerConfig
	Log      LogConfig
	Database DatabaseConfig
	TLS      TLSConfig
	WebAuthn WebAuthnConfig
	Session  SessionConfig
}

type TLSConfig struct {
	Mode     string // auto, acme, selfsigned, manual, off
	CertDir  string // Directory for auto-generated certificates
	Email    string // ACME email for Let's Encrypt
	CertFile string // Path to certificate file (manual mode)
	KeyFile  string // Path to private key file (manual mode)
}

type ServerConfig struct { //nolint:govet // fieldalignment not critical for config structs
	Host        string
	Port        int
	BaseURL     string
	MaxBodySize int // in MB
}

type LogConfig struct {
	Level  string // debug, info, warn, error
	Format string // text, json
}

type DatabaseConfig struct {
	DSN string
}

type WebAuthnConfig struct {
	RPID          string // Relying Party ID (domain), e.g. "localhost"
	RPOrigin      string // Relying Party Origin (full URL), e.g. "http://localhost:8080"
	RPDisplayName string // Display name shown to users
}

type SessionConfig struct { //nolint:govet // fieldalignment not critical
	CookieName string // Session cookie name
	MaxAge     int    // Session max age in seconds
	HashKey    string // 32-byte hex string for HMAC signing
	BlockKey   string // 32-byte hex string for AES encryption (optional)
}

func NewFromCLI(cmd *cli.Command) *Config {
	cfg := &Config{
		Server: ServerConfig{
			Host:        cmd.String("host"),
			Port:        int(cmd.Int("port")),
			BaseURL:     cmd.String("base-url"),
			MaxBodySize: int(cmd.Int("max-body-size")),
		},
		Log: LogConfig{
			Level:  cmd.String("log-level"),
			Format: cmd.String("log-format"),
		},
		Database: DatabaseConfig{
			DSN: cmd.String("database-dsn"),
		},
		TLS: TLSConfig{
			Mode:     cmd.String("tls-mode"),
			CertDir:  cmd.String("tls-cert-dir"),
			Email:    cmd.String("tls-email"),
			CertFile: cmd.String("tls-cert-file"),
			KeyFile:  cmd.String("tls-key-file"),
		},
		WebAuthn: WebAuthnConfig{
			RPID:          cmd.String("webauthn-rp-id"),
			RPOrigin:      cmd.String("webauthn-rp-origin"),
			RPDisplayName: cmd.String("webauthn-rp-display-name"),
		},
		Session: SessionConfig{
			CookieName: cmd.String("session-cookie-name"),
			MaxAge:     int(cmd.Int("session-max-age")),
			HashKey:    cmd.String("session-hash-key"),
			BlockKey:   cmd.String("session-block-key"),
		},
	}

	if cfg.Server.BaseURL == "" {
		cfg.Server.BaseURL = buildBaseURL(cfg)
	}

	// Apply WebAuthn defaults based on BaseURL
	applyWebAuthnDefaults(cfg)

	return cfg
}

// applyWebAuthnDefaults sets WebAuthn defaults based on the resolved BaseURL.
func applyWebAuthnDefaults(cfg *Config) {
	// Extract host from BaseURL for RPID
	if cfg.WebAuthn.RPID == "" {
		cfg.WebAuthn.RPID = cfg.Server.Host
	}
	// Use BaseURL as RPOrigin
	if cfg.WebAuthn.RPOrigin == "" {
		cfg.WebAuthn.RPOrigin = cfg.Server.BaseURL
	}
	// Default display name
	if cfg.WebAuthn.RPDisplayName == "" {
		cfg.WebAuthn.RPDisplayName = "Go Web App"
	}
}

func buildBaseURL(cfg *Config) string {
	host := cfg.Server.Host
	port := cfg.Server.Port
	mode := strings.ToLower(cfg.TLS.Mode)

	// Determine if TLS will be used
	useTLS := shouldUseTLS(mode, host)

	scheme := "http"
	if useTLS {
		scheme = "https"
	}

	// ACME mode always uses port 443
	if mode == "acme" {
		return fmt.Sprintf("https://%s", host)
	}

	// Hide default ports in URL
	if (scheme == "http" && port == 80) || (scheme == "https" && port == 443) {
		return fmt.Sprintf("%s://%s", scheme, host)
	}
	return fmt.Sprintf("%s://%s:%d", scheme, host, port)
}

func shouldUseTLS(mode, host string) bool {
	switch mode {
	case "off":
		return false
	case "acme", "selfsigned", "manual":
		return true
	default: // "auto" or empty
		return !IsLocalhost(host)
	}
}

// IsLocalhost checks if the host is a localhost address.
func IsLocalhost(host string) bool {
	switch host {
	case "", "localhost", "127.0.0.1", "::1":
		return true
	}
	// Check for *.localhost subdomains (e.g., app.localhost)
	return strings.HasSuffix(host, ".localhost")
}

func Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "host",
			Value:   "localhost",
			Usage:   "Host to bind to",
			Sources: cli.NewValueSourceChain(cli.EnvVar("HOST"), toml.TOML("server.host", configFile)),
		},
		&cli.IntFlag{
			Name:    "port",
			Value:   8080,
			Usage:   "Port to listen on",
			Sources: cli.NewValueSourceChain(cli.EnvVar("PORT"), toml.TOML("server.port", configFile)),
		},
		&cli.StringFlag{
			Name:    "base-url",
			Usage:   "Base URL for the application",
			Sources: cli.NewValueSourceChain(cli.EnvVar("BASE_URL"), toml.TOML("server.base_url", configFile)),
		},
		&cli.IntFlag{
			Name:    "max-body-size",
			Value:   1,
			Usage:   "Maximum request body size in MB",
			Sources: cli.NewValueSourceChain(cli.EnvVar("MAX_BODY_SIZE"), toml.TOML("server.max_body_size", configFile)),
		},
		&cli.StringFlag{
			Name:    "log-level",
			Value:   "info",
			Usage:   "Log level (debug, info, warn, error)",
			Sources: cli.NewValueSourceChain(cli.EnvVar("LOG_LEVEL"), toml.TOML("log.level", configFile)),
		},
		&cli.StringFlag{
			Name:    "log-format",
			Value:   "text",
			Usage:   "Log format (text, json)",
			Sources: cli.NewValueSourceChain(cli.EnvVar("LOG_FORMAT"), toml.TOML("log.format", configFile)),
		},
		&cli.StringFlag{
			Name:    "database-dsn",
			Value:   "./data/app.db",
			Usage:   "Database DSN",
			Sources: cli.NewValueSourceChain(cli.EnvVar("DATABASE_DSN"), toml.TOML("database.dsn", configFile)),
		},
		&cli.StringFlag{
			Name:    "tls-mode",
			Value:   "auto",
			Usage:   "TLS mode (auto, acme, selfsigned, manual, off)",
			Sources: cli.NewValueSourceChain(cli.EnvVar("TLS_MODE"), toml.TOML("tls.mode", configFile)),
		},
		&cli.StringFlag{
			Name:    "tls-cert-dir",
			Value:   "./data/certs",
			Usage:   "Directory for auto-generated certificates",
			Sources: cli.NewValueSourceChain(cli.EnvVar("TLS_CERT_DIR"), toml.TOML("tls.cert_dir", configFile)),
		},
		&cli.StringFlag{
			Name:    "tls-email",
			Usage:   "Email for ACME/Let's Encrypt registration",
			Sources: cli.NewValueSourceChain(cli.EnvVar("TLS_EMAIL"), toml.TOML("tls.email", configFile)),
		},
		&cli.StringFlag{
			Name:    "tls-cert-file",
			Usage:   "Path to TLS certificate file (manual mode)",
			Sources: cli.NewValueSourceChain(cli.EnvVar("TLS_CERT_FILE"), toml.TOML("tls.cert_file", configFile)),
		},
		&cli.StringFlag{
			Name:    "tls-key-file",
			Usage:   "Path to TLS private key file (manual mode)",
			Sources: cli.NewValueSourceChain(cli.EnvVar("TLS_KEY_FILE"), toml.TOML("tls.key_file", configFile)),
		},
		// WebAuthn flags
		&cli.StringFlag{
			Name:    "webauthn-rp-id",
			Usage:   "WebAuthn Relying Party ID (domain, defaults to host)",
			Sources: cli.NewValueSourceChain(cli.EnvVar("WEBAUTHN_RP_ID"), toml.TOML("webauthn.rp_id", configFile)),
		},
		&cli.StringFlag{
			Name:    "webauthn-rp-origin",
			Usage:   "WebAuthn Relying Party Origin (full URL, defaults to base_url)",
			Sources: cli.NewValueSourceChain(cli.EnvVar("WEBAUTHN_RP_ORIGIN"), toml.TOML("webauthn.rp_origin", configFile)),
		},
		&cli.StringFlag{
			Name:    "webauthn-rp-display-name",
			Value:   "Go Web App",
			Usage:   "WebAuthn Relying Party display name",
			Sources: cli.NewValueSourceChain(cli.EnvVar("WEBAUTHN_RP_DISPLAY_NAME"), toml.TOML("webauthn.rp_display_name", configFile)),
		},
		// Session flags
		&cli.StringFlag{
			Name:    "session-cookie-name",
			Value:   "_session",
			Usage:   "Session cookie name",
			Sources: cli.NewValueSourceChain(cli.EnvVar("SESSION_COOKIE_NAME"), toml.TOML("session.cookie_name", configFile)),
		},
		&cli.IntFlag{
			Name:    "session-max-age",
			Value:   604800, // 7 days in seconds
			Usage:   "Session max age in seconds",
			Sources: cli.NewValueSourceChain(cli.EnvVar("SESSION_MAX_AGE"), toml.TOML("session.max_age", configFile)),
		},
		&cli.StringFlag{
			Name:    "session-hash-key",
			Usage:   "Session hash key (32-byte hex, auto-generated if empty in dev)",
			Sources: cli.NewValueSourceChain(cli.EnvVar("SESSION_HASH_KEY"), toml.TOML("session.hash_key", configFile)),
		},
		&cli.StringFlag{
			Name:    "session-block-key",
			Usage:   "Session block key for encryption (32-byte hex, optional)",
			Sources: cli.NewValueSourceChain(cli.EnvVar("SESSION_BLOCK_KEY"), toml.TOML("session.block_key", configFile)),
		},
	}
}
