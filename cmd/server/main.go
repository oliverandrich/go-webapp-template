// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package main

import (
	"context"
	"fmt"
	"os"

	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	BuildTime = "unknown"
)

// sources creates a value source chain combining env vars and TOML config
func sources(envKey, tomlKey string, tomlSrc altsrc.Sourcer) cli.ValueSourceChain {
	chain := cli.EnvVars(envKey)
	chain.Chain = append(chain.Chain, toml.TOML(tomlKey, tomlSrc))
	return chain
}

func main() {
	var configFile string

	tomlSrc := altsrc.NewStringPtrSourcer(&configFile)

	cmd := &cli.Command{
		Name:    "server",
		Usage:   "A Go webapp template server",
		Version: fmt.Sprintf("%s (built %s)", Version, BuildTime),
		Flags: []cli.Flag{
			// Config file
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Value:       "config.toml",
				Usage:       "Path to configuration file",
				Destination: &configFile,
				Sources:     cli.EnvVars("CONFIG"),
			},

			// Server settings
			&cli.StringFlag{
				Name:    "host",
				Value:   "localhost",
				Usage:   "Server host",
				Sources: sources("HOST", "server.host", tomlSrc),
			},
			&cli.IntFlag{
				Name:    "port",
				Value:   8080,
				Usage:   "Server port",
				Sources: sources("PORT", "server.port", tomlSrc),
			},
			&cli.StringFlag{
				Name:    "base-url",
				Usage:   "Base URL for the application",
				Sources: sources("BASE_URL", "server.base_url", tomlSrc),
			},

			// Database
			&cli.StringFlag{
				Name:    "database",
				Value:   "./data/app.db",
				Usage:   "SQLite database path",
				Sources: sources("DATABASE", "database.path", tomlSrc),
			},

			// Authentication
			&cli.StringFlag{
				Name:    "registration-mode",
				Value:   "internal",
				Usage:   "Registration mode: internal, open, closed",
				Sources: sources("REGISTRATION_MODE", "auth.registration_mode", tomlSrc),
			},
			&cli.StringFlag{
				Name:    "session-secret",
				Usage:   "Secret key for session encryption",
				Sources: sources("SESSION_SECRET", "auth.session_secret", tomlSrc),
			},
			&cli.IntFlag{
				Name:    "session-duration",
				Value:   24,
				Usage:   "Hours for normal sessions",
				Sources: sources("SESSION_DURATION", "auth.session_duration", tomlSrc),
			},
			&cli.IntFlag{
				Name:    "remember-me-duration",
				Value:   720,
				Usage:   "Hours for 'remember me' sessions",
				Sources: sources("REMEMBER_ME_DURATION", "auth.remember_me_duration", tomlSrc),
			},
			&cli.StringFlag{
				Name:    "cookie-name",
				Value:   "session",
				Usage:   "Session cookie name",
				Sources: sources("COOKIE_NAME", "auth.cookie_name", tomlSrc),
			},
			&cli.BoolFlag{
				Name:    "cookie-secure",
				Usage:   "HTTPS only cookie",
				Sources: sources("COOKIE_SECURE", "auth.cookie_secure", tomlSrc),
			},

			// Logging
			&cli.StringFlag{
				Name:    "log-level",
				Value:   "info",
				Usage:   "Log level: debug, info, warn, error",
				Sources: sources("LOG_LEVEL", "log.level", tomlSrc),
			},
			&cli.StringFlag{
				Name:    "log-format",
				Value:   "text",
				Usage:   "Log format: text, json",
				Sources: sources("LOG_FORMAT", "log.format", tomlSrc),
			},
		},
		Action: runServer,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
