// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

// Package auth provides authentication context helpers.
package auth

import (
	"context"

	"codeberg.org/oliverandrich/go-webapp-template/internal/ctxkeys"
	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
)

// GetUser returns the authenticated user from the context, or nil if not authenticated.
func GetUser(ctx context.Context) *models.User {
	if user, ok := ctx.Value(ctxkeys.User{}).(*models.User); ok {
		return user
	}
	return nil
}

// IsAuthenticated returns true if the context has an authenticated user.
func IsAuthenticated(ctx context.Context) bool {
	return GetUser(ctx) != nil
}
