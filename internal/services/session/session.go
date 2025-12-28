// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package session

import (
	"context"
	"net/http"
	"time"

	"github.com/alexedwards/scs/gormstore"
	"github.com/alexedwards/scs/v2"
	"gorm.io/gorm"
)

// Session keys for storing user data
const (
	UserIDKey    = "user_id"
	UserEmailKey = "user_email"
	IsAdminKey   = "is_admin"
)

// SessionUser represents the user data extracted from a valid session
type SessionUser struct {
	ID      int64
	Email   string
	IsAdmin bool
}

// NewSessionManager creates a new SCS session manager with GORM storage
func NewSessionManager(db *gorm.DB, sessionDuration time.Duration, rememberMeDuration time.Duration, cookieName string, cookieSecure bool) (*scs.SessionManager, error) {
	sessionManager := scs.New()

	// Use GORM store (auto-creates sessions table if needed)
	store, err := gormstore.New(db)
	if err != nil {
		return nil, err
	}
	sessionManager.Store = store

	// Configure session behavior
	sessionManager.Lifetime = sessionDuration
	sessionManager.IdleTimeout = 0 // No idle timeout, just use Lifetime
	sessionManager.Cookie.Name = cookieName
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.Secure = cookieSecure
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Persist = false // Session cookie by default

	return sessionManager, nil
}

// Login stores user data in the session
func Login(sm *scs.SessionManager, ctx context.Context, userID int64, email string, isAdmin bool, rememberMe bool, rememberMeDuration time.Duration) error {
	// Renew token to prevent session fixation
	if err := sm.RenewToken(ctx); err != nil {
		return err
	}

	// Store user data
	sm.Put(ctx, UserIDKey, userID)
	sm.Put(ctx, UserEmailKey, email)
	sm.Put(ctx, IsAdminKey, isAdmin)

	// Set longer lifetime for "remember me"
	if rememberMe {
		sm.SetDeadline(ctx, time.Now().Add(rememberMeDuration))
	}

	return nil
}

// Logout destroys the current session
func Logout(sm *scs.SessionManager, ctx context.Context) error {
	return sm.Destroy(ctx)
}

// GetUser retrieves the authenticated user from the session
func GetUser(sm *scs.SessionManager, ctx context.Context) *SessionUser {
	userID := sm.GetInt64(ctx, UserIDKey)
	if userID == 0 {
		return nil
	}

	return &SessionUser{
		ID:      userID,
		Email:   sm.GetString(ctx, UserEmailKey),
		IsAdmin: sm.GetBool(ctx, IsAdminKey),
	}
}

// IsAuthenticated checks if a user is logged in
func IsAuthenticated(sm *scs.SessionManager, ctx context.Context) bool {
	return sm.GetInt64(ctx, UserIDKey) != 0
}

// GetUserFromRequest is a convenience helper to get user from http.Request
func GetUserFromRequest(sm *scs.SessionManager, r *http.Request) *SessionUser {
	return GetUser(sm, r.Context())
}
