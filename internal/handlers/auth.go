// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"

	"codeberg.org/oliverandrich/go-webapp-template/internal/config"
	"codeberg.org/oliverandrich/go-webapp-template/internal/htmx"
	"codeberg.org/oliverandrich/go-webapp-template/internal/services/auth"
	"codeberg.org/oliverandrich/go-webapp-template/internal/services/session"
	authtpl "codeberg.org/oliverandrich/go-webapp-template/templates/auth"
)

type AuthHandler struct {
	authService    *auth.Service
	sessionManager *scs.SessionManager
	config         *config.AuthConfig
}

func NewAuthHandler(authService *auth.Service, sessionManager *scs.SessionManager, cfg *config.AuthConfig) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		sessionManager: sessionManager,
		config:         cfg,
	}
}

// LoginPage renders the login form
func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	next := r.URL.Query().Get("next")
	authtpl.Login("", false, "", next).Render(r.Context(), w)
}

// Login handles the login form submission
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		authtpl.Login("", false, "Invalid request", "").Render(r.Context(), w)
		return
	}

	emailAddr := r.FormValue("email")
	password := r.FormValue("password")
	rememberMe := r.FormValue("remember_me") == "on"
	next := r.FormValue("next")

	user, err := h.authService.Login(r.Context(), emailAddr, password)
	if err != nil {
		errMsg := "Invalid email or password"
		if !errors.Is(err, auth.ErrInvalidCredentials) {
			errMsg = "An error occurred"
		}
		authtpl.Login(emailAddr, rememberMe, errMsg, next).Render(r.Context(), w)
		return
	}

	// Create session
	rememberMeDuration := time.Duration(h.config.RememberMeDuration) * time.Hour
	if err := session.Login(h.sessionManager, r.Context(), user.ID, user.Email, user.IsAdmin != 0, rememberMe, rememberMeDuration); err != nil {
		authtpl.Login(emailAddr, rememberMe, "An error occurred", next).Render(r.Context(), w)
		return
	}

	// Redirect to next URL or home
	redirectURL := "/"
	if next != "" && isValidNextURL(next) {
		redirectURL = next
	}
	htmx.Redirect(w, r, redirectURL)
}

// isValidNextURL validates that the next URL is safe to redirect to
func isValidNextURL(next string) bool {
	if !strings.HasPrefix(next, "/") {
		return false
	}
	if strings.HasPrefix(next, "//") {
		return false
	}
	if strings.Contains(next, "\\") {
		return false
	}
	return true
}

// RegisterPage renders the registration form
func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if !h.config.IsRegistrationEnabled() {
		authtpl.RegistrationClosed().Render(r.Context(), w)
		return
	}
	helpTexts := h.authService.PasswordValidator().GetHelpTexts()
	authtpl.Register("", nil, helpTexts).Render(r.Context(), w)
}

// Register handles the registration form submission
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if !h.config.IsRegistrationEnabled() {
		authtpl.RegistrationClosed().Render(r.Context(), w)
		return
	}

	helpTexts := h.authService.PasswordValidator().GetHelpTexts()

	if err := r.ParseForm(); err != nil {
		authtpl.Register("", []string{"Invalid request"}, helpTexts).Render(r.Context(), w)
		return
	}

	emailAddr := r.FormValue("email")
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirm")

	// Validate password confirmation
	if password != passwordConfirm {
		authtpl.Register(emailAddr, []string{"Passwords do not match"}, helpTexts).Render(r.Context(), w)
		return
	}

	_, err := h.authService.Register(r.Context(), auth.RegisterParams{
		Email:    emailAddr,
		Password: password,
	})
	if err != nil {
		var errMsgs []string

		var pwErr *auth.PasswordValidationError
		if errors.As(err, &pwErr) {
			errMsgs = pwErr.Messages()
		} else {
			switch {
			case errors.Is(err, auth.ErrUserExists):
				errMsgs = []string{"This email is already registered"}
			case errors.Is(err, auth.ErrInvalidEmail):
				errMsgs = []string{"Invalid email address"}
			case errors.Is(err, auth.ErrRegistrationClosed):
				errMsgs = []string{"Registration is currently closed"}
			default:
				errMsgs = []string{"An error occurred"}
			}
		}
		authtpl.Register(emailAddr, errMsgs, helpTexts).Render(r.Context(), w)
		return
	}

	authtpl.RegistrationSuccess().Render(r.Context(), w)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	_ = session.Logout(h.sessionManager, r.Context())
	htmx.Redirect(w, r, "/auth/login")
}
