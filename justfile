# Go Webapp Template - Development Tasks

# Disable CGO for pure-Go builds (uses modernc.org/sqlite)
export CGO_ENABLED := "0"

# Default: list available recipes
default:
    @just --list

# Generate templ files
templ:
    templ generate

# Build CSS with Tailwind
css:
    tailwindcss -i internal/assets/static/css/input.css -o internal/assets/static/css/styles.css --minify

# Bundle and hash assets with esbuild
bundle:
    #!/usr/bin/env bash
    rm -rf internal/assets/static/dist
    mkdir -p internal/assets/static/dist
    cat internal/assets/static/js/htmx.js internal/assets/static/js/webauthn.js > /tmp/app.js
    esbuild /tmp/app.js internal/assets/static/css/styles.css \
        --outdir=internal/assets/static/dist \
        --entry-names='[name].[hash]' \
        --minify \
        --metafile=internal/assets/esbuild-meta.json

# Update htmx to latest version
htmx-update:
    curl -sL "https://unpkg.com/htmx.org@2/dist/htmx.min.js" -o internal/assets/static/js/htmx.js

# Build the app binary (with embedded assets)
build: templ css bundle
    go build -o app ./cmd/app

# Run the app
run: build
    ./app

# Run with hot-reload (requires: go install github.com/air-verse/air@latest)
dev:
    air

# Run all tests (uses dev build tag to avoid embed issues)
test:
    go test -tags dev -v ./...

# Run tests with coverage
test-cover:
    go test -tags dev -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Run linter (requires: golangci-lint)
lint:
    golangci-lint run

# Format code
fmt:
    go fmt ./...
    goimports -w .

# Clean build artifacts
clean:
    rm -f app
    rm -f coverage.out coverage.html
    rm -rf tmp/
    rm -rf internal/assets/static/dist/
    rm -f internal/assets/esbuild-meta.json

# Tidy dependencies
tidy:
    go mod tidy

# Open SQLite database shell
dbshell:
    sqlite3 ./data/app.db
