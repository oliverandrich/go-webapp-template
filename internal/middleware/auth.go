// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package middleware

import (
	"context"
	"net/http"

	"github.com/oliverandrich/go-webapp-template/internal/auth"
	"github.com/oliverandrich/go-webapp-template/internal/htmx"
	"github.com/oliverandrich/go-webapp-template/internal/models"
	"github.com/oliverandrich/go-webapp-template/internal/services/session"
	"github.com/alexedwards/scs/v2"
)

// UserLoader is an interface for loading full user data
type UserLoader interface {
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
}

// LoadUser creates middleware that loads the session user into the request context
func LoadUser(sm *scs.SessionManager, userLoader UserLoader) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionUser := session.GetUserFromRequest(sm, r)
			if sessionUser != nil {
				var user *auth.User

				// Try to load full user from DB
				if userLoader != nil {
					dbUser, err := userLoader.GetUserByID(r.Context(), sessionUser.ID)
					if err == nil && dbUser != nil {
						user = &auth.User{
							ID:      dbUser.ID,
							Email:   dbUser.Email,
							IsAdmin: dbUser.IsAdmin != 0,
						}
					}
				}

				// Fallback to session data if DB load failed
				if user == nil {
					user = &auth.User{
						ID:      sessionUser.ID,
						Email:   sessionUser.Email,
						IsAdmin: sessionUser.IsAdmin,
					}
				}

				ctx := auth.SetUser(r.Context(), user)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuth middleware redirects unauthenticated users to login
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsAuthenticated(r.Context()) {
			htmx.Redirect(w, r, "/auth/login")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin middleware ensures the user is an admin
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r.Context())
		if user == nil || !user.IsAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
