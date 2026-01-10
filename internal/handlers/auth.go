// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package handlers

import (
	"encoding/binary"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"codeberg.org/oliverandrich/go-webapp-template/internal/appcontext"
	"codeberg.org/oliverandrich/go-webapp-template/internal/config"
	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"codeberg.org/oliverandrich/go-webapp-template/internal/repository"
	"codeberg.org/oliverandrich/go-webapp-template/internal/services/email"
	"codeberg.org/oliverandrich/go-webapp-template/internal/services/recovery"
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
	recovery *recovery.Service
	email    *email.Service // nil if email mode is disabled
	authCfg  *config.AuthConfig
}

// NewAuth creates a new AuthHandlers instance.
// email service can be nil if email mode is disabled.
func NewAuth(repo *repository.Repository, wa *webauthn.Service, sess *session.Manager, emailSvc *email.Service, authCfg *config.AuthConfig) *AuthHandlers {
	return &AuthHandlers{
		repo:     repo,
		webauthn: wa,
		sessions: sess,
		recovery: recovery.NewService(),
		email:    emailSvc,
		authCfg:  authCfg,
	}
}

// UseEmailMode returns true if email-based authentication is enabled.
func (h *AuthHandlers) UseEmailMode() bool {
	return h.authCfg != nil && h.authCfg.UseEmail
}

// RegisterPage renders the registration page.
func (h *AuthHandlers) RegisterPage(c echo.Context) error {
	return Render(c, http.StatusOK, authtpl.Register(h.UseEmailMode()))
}

// RegisterBeginRequest is the request body for starting registration.
type RegisterBeginRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

// RegisterBegin starts the WebAuthn registration process.
func (h *AuthHandlers) RegisterBegin(c echo.Context) error {
	var req RegisterBeginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	var user *models.User
	var createErr error
	ctx := c.Request().Context()

	if h.UseEmailMode() {
		// Email mode: validate and create user with email
		if req.Email == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "email is required"})
		}

		// Use email as display name if not provided
		if req.DisplayName == "" {
			req.DisplayName = req.Email
		}

		// Check if email already exists
		exists, err := h.repo.EmailExists(ctx, req.Email)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "database error"})
		}
		if exists {
			return c.JSON(http.StatusConflict, map[string]string{"error": "email already registered"})
		}

		// Create user with email
		user, createErr = h.repo.CreateUserWithEmail(ctx, req.Email, req.DisplayName)
		if createErr != nil {
			slog.Error("failed to create user", "error", createErr, "email", req.Email)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		}
	} else {
		// Username mode: original behavior
		if req.Username == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "username is required"})
		}

		// Use username as display name if not provided
		if req.DisplayName == "" {
			req.DisplayName = req.Username
		}

		// Check if username already exists
		exists, err := h.repo.UserExists(ctx, req.Username)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "database error"})
		}
		if exists {
			return c.JSON(http.StatusConflict, map[string]string{"error": "username already taken"})
		}

		// Create user in database
		user, createErr = h.repo.CreateUser(ctx, req.Username, req.DisplayName)
		if createErr != nil {
			slog.Error("failed to create user", "error", createErr, "username", req.Username)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		}
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

	ctx := c.Request().Context()

	// Get session data
	sessionData, err := h.webauthn.GetRegistrationSession(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "registration session expired"})
	}

	// Get user from database
	user, err := h.repo.GetUserByID(ctx, userID)
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
	if createErr := h.repo.CreateCredential(ctx, dbCred); createErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to store credential"})
	}

	// Generate recovery codes
	codes, hashes, err := h.recovery.GenerateCodes(recovery.CodeCount)
	if err != nil {
		slog.Error("failed to generate recovery codes", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate recovery codes"})
	}

	// Store recovery codes
	if createErr := h.repo.CreateRecoveryCodes(ctx, user.ID, hashes); createErr != nil {
		slog.Error("failed to store recovery codes", "error", createErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to store recovery codes"})
	}

	// Email mode: send verification email and redirect to pending page
	if h.UseEmailMode() && user.Email != nil && h.authCfg.RequireVerification {
		// Generate verification token
		plainToken, tokenHash, expiresAt, tokenErr := h.email.GenerateToken()
		if tokenErr != nil {
			slog.Error("failed to generate verification token", "error", tokenErr)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate verification token"})
		}

		// Store token
		if tokenErr = h.repo.CreateEmailVerificationToken(ctx, user.ID, tokenHash, expiresAt); tokenErr != nil {
			slog.Error("failed to store verification token", "error", tokenErr)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to store verification token"})
		}

		// Send verification email (async)
		go func() {
			if sendErr := h.email.SendVerification(ctx, *user.Email, plainToken); sendErr != nil {
				slog.Error("failed to send verification email", "error", sendErr, "email", *user.Email)
			}
		}()

		// Store codes in flash cookie for later display after verification
		flashCookie, flashErr := h.sessions.SetFlash(&session.FlashData{RecoveryCodes: codes})
		if flashErr != nil {
			slog.Error("failed to create flash cookie", "error", flashErr)
		} else {
			c.SetCookie(flashCookie)
		}

		return c.JSON(http.StatusOK, map[string]any{
			"status":   "ok",
			"redirect": "/auth/verify-pending",
		})
	}

	// Username mode or email already verified: create session immediately
	sessionCookie, err := h.sessions.Create(user.ID, user.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
	}
	c.SetCookie(sessionCookie)

	// Store codes in flash cookie for display on next page
	flashCookie, err := h.sessions.SetFlash(&session.FlashData{RecoveryCodes: codes})
	if err != nil {
		slog.Error("failed to create flash cookie", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to store recovery codes"})
	}
	c.SetCookie(flashCookie)

	return c.JSON(http.StatusOK, map[string]any{
		"status":   "ok",
		"redirect": "/auth/recovery-codes",
	})
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

	// Check email verification in email mode
	if h.UseEmailMode() && h.authCfg.RequireVerification && !foundUser.EmailVerified {
		return c.JSON(http.StatusForbidden, map[string]any{
			"error":    "email_not_verified",
			"redirect": "/auth/verify-pending",
		})
	}

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
	cc, ok := c.(*appcontext.Context)
	if !ok || !cc.IsAuthenticated() {
		return c.Redirect(http.StatusSeeOther, "/auth/login")
	}
	user := cc.GetUser()

	creds, err := h.repo.GetCredentialsByUserID(c.Request().Context(), user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get credentials"})
	}

	return Render(c, http.StatusOK, authtpl.Credentials(creds))
}

// AddCredentialBegin starts the process of adding a new credential.
func (h *AuthHandlers) AddCredentialBegin(c echo.Context) error {
	cc, ok := c.(*appcontext.Context)
	if !ok || !cc.IsAuthenticated() {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
	}
	user := cc.GetUser()

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
	cc, ok := c.(*appcontext.Context)
	if !ok || !cc.IsAuthenticated() {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
	}
	user := cc.GetUser()

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
	cc, ok := c.(*appcontext.Context)
	if !ok || !cc.IsAuthenticated() {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
	}
	user := cc.GetUser()

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

// RecoveryPage renders the recovery login page.
func (h *AuthHandlers) RecoveryPage(c echo.Context) error {
	return Render(c, http.StatusOK, authtpl.Recovery())
}

// RecoveryLoginRequest is the request body for recovery login.
type RecoveryLoginRequest struct {
	Username string `json:"username" form:"username"`
	Code     string `json:"code" form:"code"`
}

// RecoveryLogin authenticates a user with a recovery code.
func (h *AuthHandlers) RecoveryLogin(c echo.Context) error {
	var req RecoveryLoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.Username == "" || req.Code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "username and code are required"})
	}

	// Find user
	user, err := h.repo.GetUserByUsername(c.Request().Context(), req.Username)
	if err != nil {
		// Don't reveal if user exists or not
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid username or recovery code"})
	}

	// Normalize and validate recovery code
	normalizedCode := recovery.NormalizeCode(req.Code)
	valid, err := h.repo.ValidateAndUseRecoveryCode(c.Request().Context(), user.ID, normalizedCode)
	if err != nil {
		slog.Error("failed to validate recovery code", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "validation error"})
	}
	if !valid {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid username or recovery code"})
	}

	// Create session cookie
	cookie, err := h.sessions.Create(user.ID, user.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
	}
	c.SetCookie(cookie)

	// Get remaining codes count for warning
	remaining, _ := h.repo.GetUnusedRecoveryCodeCount(c.Request().Context(), user.ID)

	return c.JSON(http.StatusOK, map[string]any{
		"status":          "ok",
		"remaining_codes": remaining,
	})
}

// RegenerateRecoveryCodes generates new recovery codes and invalidates old ones.
func (h *AuthHandlers) RegenerateRecoveryCodes(c echo.Context) error {
	cc, ok := c.(*appcontext.Context)
	if !ok || !cc.IsAuthenticated() {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
	}
	user := cc.GetUser()

	// Delete old codes
	if err := h.repo.DeleteRecoveryCodes(c.Request().Context(), user.ID); err != nil {
		slog.Error("failed to delete old recovery codes", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to regenerate codes"})
	}

	// Generate new codes
	codes, hashes, err := h.recovery.GenerateCodes(recovery.CodeCount)
	if err != nil {
		slog.Error("failed to generate recovery codes", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate codes"})
	}

	// Store new codes
	if createErr := h.repo.CreateRecoveryCodes(c.Request().Context(), user.ID, hashes); createErr != nil {
		slog.Error("failed to store recovery codes", "error", createErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to store codes"})
	}

	// Store codes in flash cookie for display on next page
	flashCookie, err := h.sessions.SetFlash(&session.FlashData{RecoveryCodes: codes})
	if err != nil {
		slog.Error("failed to create flash cookie", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to store codes"})
	}
	c.SetCookie(flashCookie)

	return c.JSON(http.StatusOK, map[string]any{
		"status":   "ok",
		"redirect": "/auth/recovery-codes",
	})
}

// RecoveryCodesPage displays recovery codes from flash data.
func (h *AuthHandlers) RecoveryCodesPage(c echo.Context) error {
	// Get codes from flash cookie
	flash := h.sessions.GetFlash(c.Request())
	if flash == nil || len(flash.RecoveryCodes) == 0 {
		// No codes to display, redirect to dashboard
		return c.Redirect(http.StatusSeeOther, "/dashboard")
	}

	// Clear flash cookie
	c.SetCookie(h.sessions.ClearFlash())

	return Render(c, http.StatusOK, authtpl.RecoveryCodes(flash.RecoveryCodes))
}

// VerifyPendingPage renders the "check your email" page.
func (h *AuthHandlers) VerifyPendingPage(c echo.Context) error {
	return Render(c, http.StatusOK, authtpl.VerifyPending())
}

// VerifyEmail handles the email verification link.
func (h *AuthHandlers) VerifyEmail(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return Render(c, http.StatusBadRequest, authtpl.VerifyError("missing_token"))
	}

	ctx := c.Request().Context()

	// Hash the token to look it up
	tokenHash := email.HashToken(token)

	// Find the token
	verificationToken, err := h.repo.GetEmailVerificationToken(ctx, tokenHash)
	if err != nil {
		slog.Error("verification token not found", "error", err)
		return Render(c, http.StatusBadRequest, authtpl.VerifyError("invalid_token"))
	}

	// Check if token is expired
	if time.Now().After(verificationToken.ExpiresAt) {
		// Delete expired token
		_ = h.repo.DeleteEmailVerificationToken(ctx, verificationToken.ID)
		return Render(c, http.StatusBadRequest, authtpl.VerifyError("token_expired"))
	}

	// Mark email as verified
	if markErr := h.repo.MarkEmailVerified(ctx, verificationToken.UserID); markErr != nil {
		slog.Error("failed to mark email as verified", "error", markErr)
		return Render(c, http.StatusInternalServerError, authtpl.VerifyError("verification_failed"))
	}

	// Delete all verification tokens for this user
	_ = h.repo.DeleteUserEmailVerificationTokens(ctx, verificationToken.UserID)

	// Get user for session creation
	user, err := h.repo.GetUserByID(ctx, verificationToken.UserID)
	if err != nil {
		slog.Error("failed to get user after verification", "error", err)
		return Render(c, http.StatusInternalServerError, authtpl.VerifyError("verification_failed"))
	}

	// Create session
	sessionCookie, err := h.sessions.Create(user.ID, user.Username)
	if err != nil {
		slog.Error("failed to create session after verification", "error", err)
		return Render(c, http.StatusInternalServerError, authtpl.VerifyError("verification_failed"))
	}
	c.SetCookie(sessionCookie)

	return Render(c, http.StatusOK, authtpl.VerifySuccess())
}

// ResendVerificationRequest is the request body for resending verification email.
type ResendVerificationRequest struct {
	Email string `json:"email" form:"email"`
}

// ResendVerification resends the verification email.
func (h *AuthHandlers) ResendVerification(c echo.Context) error {
	var req ResendVerificationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "email is required"})
	}

	ctx := c.Request().Context()

	// Find user by email
	user, err := h.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if email exists
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	// Check if already verified
	if user.EmailVerified {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	// Delete existing tokens for this user
	_ = h.repo.DeleteUserEmailVerificationTokens(ctx, user.ID)

	// Generate new verification token
	plainToken, tokenHash, expiresAt, tokenErr := h.email.GenerateToken()
	if tokenErr != nil {
		slog.Error("failed to generate verification token", "error", tokenErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to send verification email"})
	}

	// Store token
	if tokenErr = h.repo.CreateEmailVerificationToken(ctx, user.ID, tokenHash, expiresAt); tokenErr != nil {
		slog.Error("failed to store verification token", "error", tokenErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to send verification email"})
	}

	// Send verification email (async)
	go func() {
		if sendErr := h.email.SendVerification(ctx, *user.Email, plainToken); sendErr != nil {
			slog.Error("failed to send verification email", "error", sendErr, "email", *user.Email)
		}
	}()

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
