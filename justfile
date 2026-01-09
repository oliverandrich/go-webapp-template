# Go Webapp Template - Development Tasks

# Default: list available recipes
default:
    @just --list

# Generate templ files
templ:
    templ generate

# Build CSS with Tailwind (hashed filename for cache busting)
css:
    #!/usr/bin/env bash
    rm -f static/css/styles.*.css
    tailwindcss -i assets/css/input.css -o static/css/styles.css --minify
    hash=$(md5sum static/css/styles.css | cut -c1-8)
    mv static/css/styles.css "static/css/styles.${hash}.css"

# Build the app binary
build: templ css
    go build -o app ./cmd/app

# Run the app
run: build
    ./app

# Run with hot-reload (requires: go install github.com/air-verse/air@latest)
dev:
    air

# Run all tests
test:
    go test -v ./...

# Run tests with coverage
test-cover:
    go test -coverprofile=coverage.out ./...
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
    rm -f static/css/styles.*.css

# Tidy dependencies
tidy:
    go mod tidy
