# Go Webapp Template

A modern, minimal Go web application template using Echo, GORM, templ, and Tailwind CSS.

## Quick Start

Create a new project using [gohatch](https://github.com/oliverandrich/gohatch):

```bash
# Install gohatch
brew tap oliverandrich/tap
brew install gohatch

# Create new project
gohatch github.com/oliverandrich/go-webapp-template github.com/example/myapp
cd myapp

# Configure and run
cp config.example.toml config.toml
just dev
```

## Features

- **[Echo](https://echo.labstack.com/)** - High performance web framework
- **[GORM](https://gorm.io/)** + SQLite - Database ORM with SQLite backend
- **[templ](https://templ.guide/)** - Type-safe HTML templating
- **[htmx](https://htmx.org/)** - High power tools for HTML (auto-downloaded, no Node.js)
- **[Tailwind CSS](https://tailwindcss.com/)** v4 - Utility-first CSS (standalone CLI, no Node.js)
- **[go-i18n](https://github.com/nicksnyder/go-i18n)** - Internationalization support
- **[urfave/cli](https://cli.urfave.org/)** - CLI with TOML configuration
- **[slog](https://pkg.go.dev/log/slog)** - Structured logging
- **Automatic TLS** - Let's Encrypt, self-signed, or manual certificates
- **WebAuthn/Passkeys** - Passwordless authentication with usernameless login
- **Email Authentication** - Optional email-based registration with verification
- CSRF protection with secure cookies
- Content-hashed static assets with immutable caching
- Custom Echo context with htmx request detection

## Project Structure

```
├── cmd/app/              # Application entrypoint
├── internal/
│   ├── config/           # Configuration handling
│   ├── ctxkeys/          # Typed context keys
│   ├── database/         # Database connection and migrations
│   ├── handlers/         # HTTP handlers
│   ├── htmx/             # htmx request parsing
│   ├── i18n/             # Internationalization
│   ├── models/           # GORM models
│   ├── repository/       # Data access layer
│   ├── server/           # Server setup, middleware, routing, custom context
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

| Setting              | Env Variable         | Default               | Description                            |
| -------------------- | -------------------- | --------------------- | -------------------------------------- |
| server.host          | HOST                 | localhost             | Bind address                           |
| server.port          | PORT                 | 8080                  | Port number                            |
| server.base_url      | BASE_URL             | (auto-generated)      | Public URL                             |
| server.max_body_size | MAX_BODY_SIZE        | 1                     | Max body size (MB)                     |
| log.level            | LOG_LEVEL            | info                  | Log level (debug/info/warn/error)      |
| log.format           | LOG_FORMAT           | text                  | Log format (text/json)                 |
| database.dsn         | DATABASE_DSN         | ./data/app.db         | SQLite path                            |
| tls.mode             | TLS_MODE             | auto                  | TLS mode (auto/acme/selfsigned/manual/off) |
| tls.cert_dir         | TLS_CERT_DIR         | ./data/certs          | Directory for auto-generated certs     |
| tls.email            | TLS_EMAIL            |                       | Email for Let's Encrypt (required for acme) |
| tls.cert_file        | TLS_CERT_FILE        |                       | Path to certificate (manual mode)      |
| tls.key_file         | TLS_KEY_FILE         |                       | Path to private key (manual mode)      |
| webauthn.rp_id       | WEBAUTHN_RP_ID       | (from host)           | WebAuthn Relying Party ID (domain)     |
| webauthn.rp_origin   | WEBAUTHN_RP_ORIGIN   | (from base_url)       | WebAuthn Relying Party Origin          |
| webauthn.rp_display_name | WEBAUTHN_RP_DISPLAY_NAME | Go Web App      | Display name for passkey prompts       |
| session.cookie_name  | SESSION_COOKIE_NAME  | _session              | Session cookie name                    |
| session.max_age      | SESSION_MAX_AGE      | 604800                | Session max age (seconds, 7 days)      |
| session.hash_key     | SESSION_HASH_KEY     | (auto in dev)         | 32-byte hex HMAC key                   |
| session.block_key    | SESSION_BLOCK_KEY    |                       | 32-byte hex AES key (optional)         |
| auth.use_email       | AUTH_USE_EMAIL       | false                 | Use email instead of username          |
| auth.require_verification | AUTH_REQUIRE_VERIFICATION | true         | Require email verification before login |
| smtp.host            | SMTP_HOST            |                       | SMTP server host                       |
| smtp.port            | SMTP_PORT            | 587                   | SMTP port (465 for TLS, 587 for STARTTLS) |
| smtp.username        | SMTP_USERNAME        |                       | SMTP username                          |
| smtp.password        | SMTP_PASSWORD        |                       | SMTP password                          |
| smtp.from            | SMTP_FROM            |                       | Sender email address                   |
| smtp.from_name       | SMTP_FROM_NAME       |                       | Sender display name                    |
| smtp.tls             | SMTP_TLS             | true                  | Enable TLS (auto-detects mode by port) |

## TLS Configuration

The server automatically configures TLS based on the environment:

| Mode | Description |
|------|-------------|
| `auto` | Automatic detection (default): localhost → no TLS, otherwise → selfsigned |
| `off` | No TLS (HTTP only) |
| `acme` | Let's Encrypt certificates (requires ports 80/443 and `TLS_EMAIL`) |
| `selfsigned` | Auto-generated ECDSA P-256 certificate (1 year validity) |
| `manual` | User-provided certificate files |

**Examples:**

```bash
# Development (localhost) - no TLS needed
just dev

# Production with Let's Encrypt
HOST=example.com TLS_MODE=acme TLS_EMAIL=admin@example.com ./app

# LAN/Internal with self-signed certificate
HOST=192.168.1.50 TLS_MODE=selfsigned ./app

# Manual certificate
TLS_MODE=manual TLS_CERT_FILE=/path/to/cert.pem TLS_KEY_FILE=/path/to/key.pem ./app
```

Self-signed certificates are stored in `$TLS_CERT_DIR/selfsigned/` and reused until they expire (30 days before expiry triggers regeneration). The SHA256 fingerprint is logged on startup for verification.

## Static Assets

CSS and JS are served with content-hash filenames for optimal caching:

- **Production** (`just build`): `styles.abc123.css`, `htmx.abc123.js` with immutable cache headers
- **Development** (`just dev`): `styles.dev.css`, `htmx.dev.js` with no-cache headers

## htmx Integration

htmx is automatically downloaded during build. Access htmx request info in handlers:

```go
func (h *Handlers) Example(c echo.Context) error {
    cc := c.(*server.Context)
    if cc.Htmx.IsHtmx {
        // Partial response for htmx
        return Render(cc, http.StatusOK, templates.Partial())
    }
    // Full page response
    return Render(cc, http.StatusOK, templates.FullPage())
}
```

Available fields: `IsHtmx`, `IsBoosted`, `CurrentURL`, `Target`, `Trigger`, `TriggerName`, `Prompt`, `IsHistoryRestore`

## WebAuthn/Passkey Authentication

Built-in passwordless authentication using WebAuthn/Passkeys:

**Features:**
- Usernameless login (browser shows available passkeys)
- Multiple passkeys per user
- Signed session cookies (no database session storage)
- Passkey management page
- Recovery codes for account recovery

**Routes:**
- `GET /auth/register` - Registration page
- `GET /auth/login` - Login page (no username needed)
- `GET /auth/credentials` - Manage passkeys (protected)
- `GET /auth/recovery-codes` - View recovery codes (protected)
- `POST /auth/logout` - Logout

### Email Mode

Enable `auth.use_email=true` to use email addresses instead of usernames:

- Users register with email address
- Verification email sent before login is allowed
- Requires SMTP configuration

**Additional routes in email mode:**
- `GET /auth/verify-email?token=...` - Email verification link
- `GET /auth/verify-pending` - "Check your inbox" page
- `POST /auth/resend-verification` - Resend verification email

**Example configuration:**
```toml
[auth]
use_email = true

[smtp]
host = "smtp.example.com"
port = 587
username = "apikey"
password = "your-api-key"
from = "noreply@example.com"
from_name = "My App"
```

**Configuration:**
```bash
# Usually auto-detected from base_url, but can be overridden:
WEBAUTHN_RP_ID=example.com
WEBAUTHN_RP_ORIGIN=https://example.com
WEBAUTHN_RP_DISPLAY_NAME="My App"

# Session (auto-generated key in development):
SESSION_COOKIE_NAME=_session
SESSION_MAX_AGE=604800  # 7 days
SESSION_HASH_KEY=<32-byte-hex>  # Generate with: openssl rand -hex 32
```

**Access user in handlers:**
```go
func (h *Handlers) Dashboard(c echo.Context) error {
    user := auth.GetUser(c.Request().Context())
    if user != nil {
        // User is logged in
    }
    return Render(c, http.StatusOK, templates.Dashboard(user))
}
```

## License

[EUPL-1.2](LICENSE)
