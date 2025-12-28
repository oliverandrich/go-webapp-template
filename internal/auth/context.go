// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

// Package auth provides authentication context helpers that can be used
// across packages without causing import cycles.
package auth

import (
	"context"
)

// Context key type to avoid collisions
type contextKey string

// UserContextKey is the context key for the authenticated user
const UserContextKey contextKey = "user"

// User represents the authenticated user stored in context
type User struct {
	ID      int64
	Email   string
	IsAdmin bool
}

// GetUser retrieves the authenticated user from the context
func GetUser(ctx context.Context) *User {
	user, ok := ctx.Value(UserContextKey).(*User)
	if !ok {
		return nil
	}
	return user
}

// IsAuthenticated checks if a user is authenticated in the context
func IsAuthenticated(ctx context.Context) bool {
	return GetUser(ctx) != nil
}

// SetUser adds a user to the context
func SetUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}
