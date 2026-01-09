# Go Webapp Template

A modern, minimal Go web application template using Echo, GORM, templ, and Tailwind CSS.

## Features

- **[Echo](https://echo.labstack.com/)** - High performance web framework
- **[GORM](https://gorm.io/)** + SQLite - Database ORM with SQLite backend
- **[templ](https://templ.guide/)** - Type-safe HTML templating
- **[Tailwind CSS](https://tailwindcss.com/)** v4 - Utility-first CSS (standalone CLI, no Node.js)
- **[go-i18n](https://github.com/nicksnyder/go-i18n)** - Internationalization support
- **[urfave/cli](https://cli.urfave.org/)** - CLI with TOML configuration
- **[slog](https://pkg.go.dev/log/slog)** - Structured logging
- CSRF protection with secure cookies
- Content-hashed static assets with immutable caching

## Project Structure

```
├── cmd/app/              # Application entrypoint
├── internal/
│   ├── config/           # Configuration handling
│   ├── database/         # Database connection and migrations
│   ├── handlers/         # HTTP handlers
│   ├── i18n/             # Internationalization
│   ├── models/           # GORM models
│   ├── repository/       # Data access layer
│   ├── server/           # Server setup, middleware, routing
│   └── templates/        # templ templates and helpers
├── assets/
│   └── css/input.css     # Tailwind CSS input
├── static/               # Generated static files (gitignored)
├── locales/              # Translation files
├── config.example.toml   # Example configuration
└── justfile              # Development tasks
```

## Prerequisites

- Go 1.25+
- [templ](https://templ.guide/quick-start/installation) - `go install github.com/a-h/templ/cmd/templ@latest`
- [Tailwind CSS CLI](https://tailwindcss.com/blog/standalone-cli) - via Homebrew: `brew install tailwindcss`
- [Air](https://github.com/air-verse/air) (optional, for hot-reload) - `go install github.com/air-verse/air@latest`
- [just](https://github.com/casey/just) (optional, task runner) - `brew install just`

## Getting Started

1. Clone and configure:
   ```bash
   cp config.example.toml config.toml
   ```

2. Build and run:
   ```bash
   just build
   ./app
   ```

3. Or use hot-reload for development:
   ```bash
   just dev
   ```

4. Open http://localhost:8080

## Development Commands

```bash
just          # List all available commands
just build    # Generate templ + CSS, build binary
just run      # Build and run
just dev      # Hot-reload with Air
just test     # Run tests
just lint     # Run linter
just fmt      # Format code
just clean    # Clean build artifacts
just tidy     # Tidy Go modules
```

## Configuration

Configuration via `config.toml` or environment variables:

| Setting              | Env Variable         | Default               | Description                       |
| -------------------- | -------------------- | --------------------- | --------------------------------- |
| server.host          | SERVER_HOST          | localhost             | Bind address                      |
| server.port          | SERVER_PORT          | 8080                  | Port number                       |
| server.base_url      | SERVER_BASE_URL      | http://localhost:8080 | Public URL                        |
| server.max_body_size | SERVER_MAX_BODY_SIZE | 1                     | Max body size (MB)                |
| log.level            | LOG_LEVEL            | info                  | Log level (debug/info/warn/error) |
| log.format           | LOG_FORMAT           | text                  | Log format (text/json)            |
| database.dsn         | DATABASE_DSN         | ./data/app.db         | SQLite path                       |

## Static Assets

CSS is built with Tailwind CSS and served with content-hash filenames for optimal caching:

- **Production** (`just build`): `styles.abc123.css` with immutable cache headers
- **Development** (`just dev`): `styles.dev.css` with no-cache headers

## License

[EUPL-1.2](LICENSE)
