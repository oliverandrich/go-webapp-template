# Go Webapp Template - Task Runner

# Template module path (updated by setup after gonew)
template_module := "codeberg.org/oliverandrich/go-webapp-template"

# Default: show available commands
default:
    @just --list

# Initialize project: update module paths (after gonew) and install dependencies
setup:
    #!/usr/bin/env bash
    set -euo pipefail
    current_module=$(head -1 go.mod | awk '{print $2}')
    if [ "$current_module" != "{{template_module}}" ]; then
        echo "Updating module path: {{template_module}} → $current_module"
        find . -name "*.templ" -exec sed -i '' "s|{{template_module}}|$current_module|g" {} \;
    fi
    go mod download
    just htmx
    just templ
    just css
    echo "Setup complete. Run 'just dev' to start development server."

# Download HTMX and extensions
htmx:
    curl -sL https://unpkg.com/htmx.org@2/dist/htmx.min.js -o static/js/htmx.min.js
    curl -sL https://unpkg.com/htmx-ext-sse@2/sse.js -o static/js/htmx-sse.js
    @echo "Downloaded htmx.min.js and htmx-sse.js (latest 2.x)"

# Run development server with live reload
dev:
    air

# Build for production
build:
    just templ
    just css-prod
    go build -ldflags="-s -w -X 'main.Version={{version}}' -X 'main.BuildTime={{build_time}}'" -trimpath -o bin/server ./cmd/server

# Version from git (tag or short commit hash)
version := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
build_time := `date -u '+%Y-%m-%d %H:%M:%S UTC'`

# Run tests
test *ARGS:
    go test ./... {{ARGS}}

# Generate Templ templates
templ:
    templ generate

# Build Tailwind CSS
css:
    tailwindcss -i static/css/input.css -o static/css/output.css

# Build Tailwind CSS in watch mode
css-watch:
    tailwindcss -i static/css/input.css -o static/css/output.css --watch

# Build minified Tailwind CSS for production
css-prod:
    tailwindcss -i static/css/input.css -o static/css/output.css --minify

# Format code
fmt:
    go fmt ./...
    templ fmt .

# Lint code
lint:
    go vet ./...

# Run the server directly (without live reload)
run:
    go run ./cmd/server

# Clean build artifacts
clean:
    rm -rf bin/
    rm -f static/css/output.css
