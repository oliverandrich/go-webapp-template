// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

// Package ctxkeys defines typed context keys used across packages.
package ctxkeys

// CSRFToken is the context key for the CSRF token.
type CSRFToken struct{}

// CSSPath is the context key for the CSS path.
type CSSPath struct{}

// JSPath is the context key for the JS (htmx) path.
type JSPath struct{}

// User is the context key for the authenticated user.
type User struct{}
