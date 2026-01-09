// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package config

import (
	"fmt"

	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
)

var configFile = altsrc.StringSourcer("config.toml")

type Config struct { //nolint:govet // fieldalignment not critical for config structs
	Server   ServerConfig
	Log      LogConfig
	Database DatabaseConfig
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
	}

	if cfg.Server.BaseURL == "" {
		cfg.Server.BaseURL = fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
	}

	return cfg
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
	}
}
