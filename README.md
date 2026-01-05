# Go Webapp Template

A ready-to-use template for Go web applications with authentication, sessions, and i18n.

## Stack

- **Go** with **Chi v5** router
- **GORM** with **mattn/go-sqlite3** (CGO)
- **SCS** sessions with **gormstore**
- **nosurf** CSRF protection
- **go-i18n** internationalization (EN/DE)
- **urfave/cli v3** with **altsrc** (TOML/ENV/CLI config)
- **Templ** type-safe templates
- **HTMX** frontend interactivity
- **Tailwind CSS v4**

## Features

- User authentication (login, register, logout)
- Three registration modes: `open`, `internal`, `closed`
- Session management with "remember me"
- Password validation (length, common passwords, similarity check)
- i18n support (English, German)
- HTMX-ready with CSRF protection

## Quick Start

```bash
# Create new project from template
gohatch -e templ codeberg.org/oliverandrich/go-webapp-template example.com/you/your-app

# Install dependencies
cd your-app
just setup

# Start development server (with hot reload)
just dev
```

## Requirements

- Go 1.24+
- C compiler (gcc or clang) for CGO
- [gohatch](https://codeberg.org/oliverandrich/gohatch)
- [just](https://github.com/casey/just) (command runner)
- [air](https://github.com/air-verse/air) (hot reload)
- [templ](https://templ.guide)
- Node.js (for Tailwind CSS)

## Configuration

Configuration priority: CLI flags > Environment variables > TOML file

```bash
# Use config file
./bin/server --config config.toml

# Or environment variables
export PORT=3000
export DATABASE=./data/myapp.db
./bin/server

# Or CLI flags
./bin/server --port 3000 --database ./data/myapp.db
```

### config.toml

```toml
[server]
host = "localhost"
port = 8080

[database]
path = "./data/app.db"

[auth]
registration_mode = "internal"  # open | internal | closed
session_duration = 24           # hours
remember_me_duration = 720      # hours (30 days)
cookie_name = "session"
cookie_secure = false           # true for HTTPS

[log]
level = "info"                  # debug | info | warn | error
format = "text"                 # text | json
```

### Registration Modes

| Mode       | Description                                         |
| ---------- | --------------------------------------------------- |
| `open`     | Anyone can register (public apps)                   |
| `internal` | Registration without verification (LAN/family/team) |
| `closed`   | Only admin can create users                         |

## Development

```bash
just setup    # Update module paths + install dependencies (run once)
just dev      # Start dev server with hot reload
just build    # Build production binary
just test     # Run tests
just templ    # Generate templ files
just css      # Build Tailwind CSS
```

## Project Structure

```
├── cmd/server/          # Application entry point
│   ├── main.go          # CLI setup
│   ├── server.go        # Server initialization
│   └── routes.go        # Router & middleware
├── internal/
│   ├── config/          # Configuration
│   ├── database/        # GORM + SQLite setup
│   ├── models/          # Database models
│   ├── repository/      # Data access layer
│   ├── services/
│   │   ├── auth/        # Authentication logic
│   │   └── session/     # Session management
│   ├── handlers/        # HTTP handlers
│   ├── middleware/      # HTTP middleware
│   ├── i18n/            # Internationalization
│   ├── csrf/            # CSRF protection
│   └── htmx/            # HTMX helpers
├── templates/           # Templ templates
│   ├── layouts/         # Page layouts
│   ├── components/      # Reusable components
│   ├── auth/            # Auth pages
│   └── home/            # Home page
├── static/              # Static assets
│   ├── css/             # Tailwind CSS
│   └── js/              # HTMX
└── data/                # SQLite database (gitignored)
```

## Deployment

### Binary

```bash
# Build
just build

# Run (uses ./data/app.db by default)
./bin/server --cookie-secure=true
```

For production, use a reverse proxy (nginx, Caddy) for TLS termination.

## CI/CD

This template includes Woodpecker CI configuration:

- `.woodpecker/ci.yml` - Runs on push/PR: generates templates, builds CSS, runs tests

## License

EUPL-1.2 - see [LICENSE](LICENSE)
