// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package handlers

import (
	"encoding/binary"
	"log/slog"
	"net/http"
	"strconv"

	"codeberg.org/oliverandrich/go-webapp-template/internal/auth"
	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"codeberg.org/oliverandrich/go-webapp-template/internal/repository"
	"codeberg.org/oliverandrich/go-webapp-template/internal/services/session"
	"codeberg.org/oliverandrich/go-webapp-template/internal/services/webauthn"
	authtpl "codeberg.org/oliverandrich/go-webapp-template/internal/templates/auth"
	gowebauthn "github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AuthHandlers contains handlers for authentication.
type AuthHandlers struct {
	repo     *repository.Repository
	webauthn *webauthn.Service
	sessions *session.Manager
}

// NewAuth creates a new AuthHandlers instance.
func NewAuth(repo *repository.Repository, wa *webauthn.Service, sess *session.Manager) *AuthHandlers {
	return &AuthHandlers{
		repo:     repo,
		webauthn: wa,
		sessions: sess,
	}
}

// RegisterPage renders the registration page.
func (h *AuthHandlers) RegisterPage(c echo.Context) error {
	return Render(c, http.StatusOK, authtpl.Register())
}

// RegisterBeginRequest is the request body for starting registration.
type RegisterBeginRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

// RegisterBegin starts the WebAuthn registration process.
func (h *AuthHandlers) RegisterBegin(c echo.Context) error {
	var req RegisterBeginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.Username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "username is required"})
	}

	// Use username as display name if not provided
	if req.DisplayName == "" {
		req.DisplayName = req.Username
	}

	// Check if username already exists
	exists, err := h.repo.UserExists(c.Request().Context(), req.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "database error"})
	}
	if exists {
		return c.JSON(http.StatusConflict, map[string]string{"error": "username already taken"})
	}

	// Create user in database
	user, err := h.repo.CreateUser(c.Request().Context(), req.Username, req.DisplayName)
	if err != nil {
		slog.Error("failed to create user", "error", err, "username", req.Username)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
	}

	// Begin WebAuthn registration
	options, sessionData, err := h.webauthn.WebAuthn().BeginRegistration(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to begin registration"})
	}

	// Store session data
	h.webauthn.StoreRegistrationSession(user.ID, sessionData)

	return c.JSON(http.StatusOK, map[string]any{
		"publicKey": options.Response,
		"user_id":   user.ID,
	})
}

// RegisterFinishRequest is the request body for finishing registration.
type RegisterFinishRequest struct {
	UserID int64 `json:"user_id"`
}

// RegisterFinish completes the WebAuthn registration process.
func (h *AuthHandlers) RegisterFinish(c echo.Context) error {
	userID, err := strconv.ParseInt(c.QueryParam("user_id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user_id"})
	}

	// Get session data
	sessionData, err := h.webauthn.GetRegistrationSession(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "registration session expired"})
	}

	// Get user from database
	user, err := h.repo.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
	}

	// Finish registration
	credential, err := h.webauthn.WebAuthn().FinishRegistration(user, *sessionData, c.Request())
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "registration failed: " + err.Error()})
	}

	// Store credential in database
	dbCred := &models.Credential{
		UserID:          user.ID,
		CredentialID:    credential.ID,
		PublicKey:       credential.PublicKey,
		AAGUID:          credential.Authenticator.AAGUID,
		SignCount:       credential.Authenticator.SignCount,
		Transports:      models.TransportsFromWebAuthn(credential.Transport),
		Name:            "Passkey",
		BackupEligible:  credential.Flags.BackupEligible,
		BackupState:     credential.Flags.BackupState,
		AttestationType: credential.AttestationType,
	}
	if createErr := h.repo.CreateCredential(c.Request().Context(), dbCred); createErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to store credential"})
	}

	// Create session cookie
	cookie, err := h.sessions.Create(user.ID, user.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
	}
	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// LoginPage renders the login page.
func (h *AuthHandlers) LoginPage(c echo.Context) error {
	return Render(c, http.StatusOK, authtpl.Login())
}

// LoginBegin starts the WebAuthn login process (usernameless/discoverable).
func (h *AuthHandlers) LoginBegin(c echo.Context) error {
	// Begin discoverable login (no user info needed)
	options, sessionData, err := h.webauthn.WebAuthn().BeginDiscoverableLogin()
	if err != nil {
		slog.Error("failed to begin discoverable login", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to begin login"})
	}

	// Generate session ID for this login attempt
	sessionID := uuid.New().String()
	h.webauthn.StoreDiscoverableSession(sessionID, sessionData)

	return c.JSON(http.StatusOK, map[string]any{
		"publicKey":  options.Response,
		"session_id": sessionID,
	})
}

// LoginFinish completes the WebAuthn login process.
func (h *AuthHandlers) LoginFinish(c echo.Context) error {
	sessionID := c.QueryParam("session_id")
	if sessionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "session_id is required"})
	}

	// Get session data
	sessionData, err := h.webauthn.GetDiscoverableSession(sessionID)
	if err != nil {
		slog.Error("failed to get discoverable session", "error", err, "session_id", sessionID)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "login session expired"})
	}

	// Finish discoverable login with user handler
	var foundUser *models.User
	credential, finishErr := h.webauthn.WebAuthn().FinishDiscoverableLogin(
		func(rawID, userHandle []byte) (gowebauthn.User, error) {
			// userHandle contains the user ID we set during registration
			slog.Debug("discoverable login callback", "rawID_len", len(rawID), "userHandle_len", len(userHandle))
			if len(userHandle) < 8 {
				slog.Error("invalid user handle length", "length", len(userHandle))
				return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid user handle")
			}
			userID := int64(binary.BigEndian.Uint64(userHandle)) //nolint:gosec // user IDs are always positive
			slog.Debug("looking up user", "user_id", userID)
			user, userErr := h.repo.GetUserByID(c.Request().Context(), userID)
			if userErr != nil {
				slog.Error("failed to get user by ID", "error", userErr, "user_id", userID)
				return nil, userErr
			}
			foundUser = user
			return user, nil
		},
		*sessionData,
		c.Request(),
	)
	if finishErr != nil {
		slog.Error("failed to finish discoverable login", "error", finishErr)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "login failed"})
	}

	// Update sign count
	_ = h.repo.UpdateCredentialSignCount(c.Request().Context(), credential.ID, credential.Authenticator.SignCount)

	// Create session cookie
	cookie, err := h.sessions.Create(foundUser.ID, foundUser.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
	}
	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// Logout clears the session cookie.
func (h *AuthHandlers) Logout(c echo.Context) error {
	c.SetCookie(h.sessions.Clear())
	return c.Redirect(http.StatusSeeOther, "/")
}

// CredentialsPage renders the credentials management page.
func (h *AuthHandlers) CredentialsPage(c echo.Context) error {
	user := auth.GetUser(c.Request().Context())
	if user == nil {
		return c.Redirect(http.StatusSeeOther, "/auth/login")
	}

	creds, err := h.repo.GetCredentialsByUserID(c.Request().Context(), user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get credentials"})
	}

	return Render(c, http.StatusOK, authtpl.Credentials(creds))
}

// AddCredentialBegin starts the process of adding a new credential.
func (h *AuthHandlers) AddCredentialBegin(c echo.Context) error {
	user := auth.GetUser(c.Request().Context())
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
	}

	// Begin registration for existing user
	options, sessionData, err := h.webauthn.WebAuthn().BeginRegistration(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to begin registration"})
	}

	h.webauthn.StoreRegistrationSession(user.ID, sessionData)

	return c.JSON(http.StatusOK, map[string]any{
		"publicKey": options.Response,
	})
}

// AddCredentialFinish completes adding a new credential.
func (h *AuthHandlers) AddCredentialFinish(c echo.Context) error {
	user := auth.GetUser(c.Request().Context())
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
	}

	// Get session data
	sessionData, err := h.webauthn.GetRegistrationSession(user.ID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "registration session expired"})
	}

	// Finish registration
	credential, err := h.webauthn.WebAuthn().FinishRegistration(user, *sessionData, c.Request())
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "registration failed: " + err.Error()})
	}

	// Store credential
	dbCred := &models.Credential{
		UserID:          user.ID,
		CredentialID:    credential.ID,
		PublicKey:       credential.PublicKey,
		AAGUID:          credential.Authenticator.AAGUID,
		SignCount:       credential.Authenticator.SignCount,
		Transports:      models.TransportsFromWebAuthn(credential.Transport),
		Name:            "Passkey",
		BackupEligible:  credential.Flags.BackupEligible,
		BackupState:     credential.Flags.BackupState,
		AttestationType: credential.AttestationType,
	}
	if err := h.repo.CreateCredential(c.Request().Context(), dbCred); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to store credential"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// DeleteCredential removes a credential.
func (h *AuthHandlers) DeleteCredential(c echo.Context) error {
	user := auth.GetUser(c.Request().Context())
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
	}

	credID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid credential id"})
	}

	// Check if this is the last credential
	count, err := h.repo.CountUserCredentials(c.Request().Context(), user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "database error"})
	}
	if count <= 1 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "cannot delete last credential"})
	}

	// Delete credential
	if err := h.repo.DeleteCredential(c.Request().Context(), credID, user.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete credential"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
