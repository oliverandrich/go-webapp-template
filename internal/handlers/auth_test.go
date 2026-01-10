// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"codeberg.org/oliverandrich/go-webapp-template/internal/appcontext"
	"codeberg.org/oliverandrich/go-webapp-template/internal/config"
	"codeberg.org/oliverandrich/go-webapp-template/internal/handlers"
	"codeberg.org/oliverandrich/go-webapp-template/internal/i18n"
	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"codeberg.org/oliverandrich/go-webapp-template/internal/repository"
	"codeberg.org/oliverandrich/go-webapp-template/internal/services/session"
	"codeberg.org/oliverandrich/go-webapp-template/internal/services/webauthn"
	"codeberg.org/oliverandrich/go-webapp-template/internal/testutil"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

// newTestContext creates an appcontext.Context for testing.
func newTestContext(e *echo.Echo, req *http.Request, rec *httptest.ResponseRecorder, user *models.User) *appcontext.Context {
	return &appcontext.Context{
		Context: e.NewContext(req, rec),
		User:    user,
	}
}

// validHashKey for session manager in tests
const testHashKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func newTestAuthHandlers(t *testing.T) (*handlers.AuthHandlers, *repository.Repository) {
	t.Helper()
	db := testutil.NewTestDB(t)
	repo := repository.New(db)

	waSvc, err := webauthn.NewService(&config.WebAuthnConfig{
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
		RPDisplayName: "Test App",
	})
	require.NoError(t, err)

	sessMgr, err := session.NewManager(&config.SessionConfig{
		CookieName: "_test_session",
		MaxAge:     3600,
		HashKey:    testHashKey,
	}, false)
	require.NoError(t, err)

	// Use nil email service and default auth config (username mode)
	h := handlers.NewAuth(repo, waSvc, sessMgr, nil, &config.AuthConfig{UseEmail: false})
	return h, repo
}

func TestNewAuth(t *testing.T) {
	h, _ := newTestAuthHandlers(t)
	assert.NotNil(t, h)
}

func TestRegisterPage(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/auth/register", nil)
	ctx := i18n.WithLocale(req.Context(), language.English)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.RegisterPage(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<!doctype html>")
}

func TestRegisterBegin(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	body := strings.NewReader(`{"username":"newuser"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register/begin", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.RegisterBegin(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "publicKey")
	assert.Contains(t, rec.Body.String(), "user_id")
}

func TestRegisterBegin_MissingUsername(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	body := strings.NewReader(`{"username":""}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register/begin", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.RegisterBegin(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "username is required")
}

func TestRegisterBegin_UsernameExists(t *testing.T) {
	h, repo := newTestAuthHandlers(t)

	// Create existing user
	_, err := repo.CreateUser(context.Background(), "existinguser")
	require.NoError(t, err)

	e := echo.New()
	body := strings.NewReader(`{"username":"existinguser"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register/begin", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.RegisterBegin(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "username already taken")
}

func TestRegisterFinish_InvalidUserID(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/register/finish?user_id=invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.RegisterFinish(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid user_id")
}

func TestRegisterFinish_SessionExpired(t *testing.T) {
	h, repo := newTestAuthHandlers(t)

	// Create user but don't store registration session
	user := testutil.NewTestUser(t, repo.DB(), "testuser")

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/register/finish?user_id="+string(rune(user.ID+'0')), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/auth/register/finish")

	err := h.RegisterFinish(c)

	require.NoError(t, err)
	// Either session expired or user not found, both are expected
	assert.True(t, rec.Code == http.StatusBadRequest || rec.Code == http.StatusNotFound)
}

func TestLoginPage(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	ctx := i18n.WithLocale(req.Context(), language.English)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.LoginPage(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<!doctype html>")
}

func TestLoginBegin(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/login/begin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.LoginBegin(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "publicKey")
	assert.Contains(t, rec.Body.String(), "session_id")
}

func TestLoginFinish_MissingSessionID(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/login/finish", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.LoginFinish(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "session_id is required")
}

func TestLoginFinish_SessionExpired(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/login/finish?session_id=nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.LoginFinish(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "login session expired")
}

func TestLogout(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Logout(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/", rec.Header().Get("Location"))

	// Should have cleared the cookie
	cookies := rec.Result().Cookies()
	require.NotEmpty(t, cookies)
	assert.Equal(t, "_test_session", cookies[0].Name)
	assert.Equal(t, -1, cookies[0].MaxAge)
}

func TestCredentialsPage_Unauthenticated(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/auth/credentials", nil)
	rec := httptest.NewRecorder()
	c := newTestContext(e, req, rec, nil)

	err := h.CredentialsPage(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/auth/login", rec.Header().Get("Location"))
}

func TestCredentialsPage_Authenticated(t *testing.T) {
	h, repo := newTestAuthHandlers(t)

	user := testutil.NewTestUser(t, repo.DB(), "testuser")
	testutil.NewTestCredential(t, repo.DB(), user.ID, "cred-1")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/auth/credentials", nil)
	ctx := i18n.WithLocale(req.Context(), language.English)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := newTestContext(e, req, rec, user)

	err := h.CredentialsPage(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<!doctype html>")
}

func TestAddCredentialBegin_Unauthenticated(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/credentials/begin", nil)
	rec := httptest.NewRecorder()
	c := newTestContext(e, req, rec, nil)

	err := h.AddCredentialBegin(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "not authenticated")
}

func TestAddCredentialBegin_Authenticated(t *testing.T) {
	h, repo := newTestAuthHandlers(t)

	user := testutil.NewTestUser(t, repo.DB(), "testuser")

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/credentials/begin", nil)
	rec := httptest.NewRecorder()
	c := newTestContext(e, req, rec, user)

	err := h.AddCredentialBegin(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "publicKey")
}

func TestAddCredentialFinish_Unauthenticated(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/credentials/finish", nil)
	rec := httptest.NewRecorder()
	c := newTestContext(e, req, rec, nil)

	err := h.AddCredentialFinish(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAddCredentialFinish_SessionExpired(t *testing.T) {
	h, repo := newTestAuthHandlers(t)

	user := testutil.NewTestUser(t, repo.DB(), "testuser")

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/credentials/finish", nil)
	rec := httptest.NewRecorder()
	c := newTestContext(e, req, rec, user)

	err := h.AddCredentialFinish(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "registration session expired")
}

func TestDeleteCredential_Unauthenticated(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/auth/credentials/1", nil)
	rec := httptest.NewRecorder()
	c := newTestContext(e, req, rec, nil)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := h.DeleteCredential(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestDeleteCredential_InvalidID(t *testing.T) {
	h, repo := newTestAuthHandlers(t)

	user := testutil.NewTestUser(t, repo.DB(), "testuser")

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/auth/credentials/invalid", nil)
	rec := httptest.NewRecorder()
	c := newTestContext(e, req, rec, user)
	c.SetParamNames("id")
	c.SetParamValues("invalid")

	err := h.DeleteCredential(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid credential id")
}

func TestDeleteCredential_LastCredential(t *testing.T) {
	h, repo := newTestAuthHandlers(t)

	user := testutil.NewTestUser(t, repo.DB(), "testuser")
	cred := testutil.NewTestCredential(t, repo.DB(), user.ID, "only-cred")

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/auth/credentials/"+string(rune(cred.ID+'0')), nil)
	rec := httptest.NewRecorder()
	c := newTestContext(e, req, rec, user)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := h.DeleteCredential(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "cannot delete last credential")
}

func TestDeleteCredential_Success(t *testing.T) {
	h, repo := newTestAuthHandlers(t)

	user := testutil.NewTestUser(t, repo.DB(), "testuser")
	cred1 := testutil.NewTestCredential(t, repo.DB(), user.ID, "cred-1")
	testutil.NewTestCredential(t, repo.DB(), user.ID, "cred-2")

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/auth/credentials/1", nil)
	rec := httptest.NewRecorder()
	c := newTestContext(e, req, rec, user)
	c.SetParamNames("id")
	c.SetParamValues(strconv.FormatInt(cred1.ID, 10))

	err := h.DeleteCredential(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRegisterBegin_UsernameOnly(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	body := strings.NewReader(`{"username":"testuser"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register/begin", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.RegisterBegin(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "publicKey")
}

func TestRegisterBegin_InvalidJSON(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	body := strings.NewReader(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register/begin", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.RegisterBegin(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid request")
}

func TestRegisterFinish_NoUserID(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/register/finish", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.RegisterFinish(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid user_id")
}

func TestRegisterFinish_UserNotFound(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	// Use a valid user_id format but non-existent user
	req := httptest.NewRequest(http.MethodPost, "/auth/register/finish?user_id=99999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.RegisterFinish(c)

	require.NoError(t, err)
	// Should get session expired since we didn't store a registration session
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLoginFinish_EmptySessionID(t *testing.T) {
	h, _ := newTestAuthHandlers(t)

	e := echo.New()
	// Explicitly set empty session_id
	req := httptest.NewRequest(http.MethodPost, "/auth/login/finish?session_id=", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.LoginFinish(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "session_id is required")
}

func TestAddCredentialBegin_Success(t *testing.T) {
	h, repo := newTestAuthHandlers(t)

	user := testutil.NewTestUser(t, repo.DB(), "testuser")

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/credentials/begin", nil)
	rec := httptest.NewRecorder()
	c := newTestContext(e, req, rec, user)

	err := h.AddCredentialBegin(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "publicKey")
}

func TestCredentialsPage_NoCredentials(t *testing.T) {
	h, repo := newTestAuthHandlers(t)

	user := testutil.NewTestUser(t, repo.DB(), "testuser")
	// Don't add any credentials

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/auth/credentials", nil)
	ctx := i18n.WithLocale(req.Context(), language.English)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := newTestContext(e, req, rec, user)

	err := h.CredentialsPage(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<!doctype html>")
}

// Email mode tests

func newTestEmailAuthHandlers(t *testing.T) (*handlers.AuthHandlers, *repository.Repository) {
	t.Helper()
	db := testutil.NewTestDB(t)
	repo := repository.New(db)

	waSvc, err := webauthn.NewService(&config.WebAuthnConfig{
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
		RPDisplayName: "Test App",
	})
	require.NoError(t, err)

	sessMgr, err := session.NewManager(&config.SessionConfig{
		CookieName: "_test_session",
		MaxAge:     3600,
		HashKey:    testHashKey,
	}, false)
	require.NoError(t, err)

	// Email mode enabled, but without email service (for unit testing handlers)
	h := handlers.NewAuth(repo, waSvc, sessMgr, nil, &config.AuthConfig{
		UseEmail:            true,
		RequireVerification: true,
	})
	return h, repo
}

func TestUseEmailMode_Enabled(t *testing.T) {
	h, _ := newTestEmailAuthHandlers(t)
	assert.True(t, h.UseEmailMode())
}

func TestUseEmailMode_Disabled(t *testing.T) {
	h, _ := newTestAuthHandlers(t)
	assert.False(t, h.UseEmailMode())
}

func TestRegisterBegin_EmailMode_Success(t *testing.T) {
	h, _ := newTestEmailAuthHandlers(t)

	e := echo.New()
	body := strings.NewReader(`{"email":"test@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register/begin", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.RegisterBegin(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "publicKey")
	assert.Contains(t, rec.Body.String(), "user_id")
}

func TestRegisterBegin_EmailMode_MissingEmail(t *testing.T) {
	h, _ := newTestEmailAuthHandlers(t)

	e := echo.New()
	body := strings.NewReader(`{"email":""}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register/begin", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.RegisterBegin(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "email is required")
}

func TestRegisterBegin_EmailMode_EmailExists(t *testing.T) {
	h, repo := newTestEmailAuthHandlers(t)

	// Create existing user with email
	_, err := repo.CreateUserWithEmail(context.Background(), "existing@example.com")
	require.NoError(t, err)

	e := echo.New()
	body := strings.NewReader(`{"email":"existing@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register/begin", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.RegisterBegin(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "email already registered")
}

func TestRegisterBegin_EmailMode_EmailOnly(t *testing.T) {
	h, _ := newTestEmailAuthHandlers(t)

	e := echo.New()
	body := strings.NewReader(`{"email":"test@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register/begin", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.RegisterBegin(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "publicKey")
}

func TestVerifyPendingPage(t *testing.T) {
	h, _ := newTestEmailAuthHandlers(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/auth/verify-pending", nil)
	ctx := i18n.WithLocale(req.Context(), language.English)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.VerifyPendingPage(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<!doctype html>")
}

func TestVerifyEmail_MissingToken(t *testing.T) {
	h, _ := newTestEmailAuthHandlers(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/auth/verify-email", nil)
	ctx := i18n.WithLocale(req.Context(), language.English)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.VerifyEmail(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "<!doctype html>")
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	h, _ := newTestEmailAuthHandlers(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/auth/verify-email?token=invalid", nil)
	ctx := i18n.WithLocale(req.Context(), language.English)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.VerifyEmail(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "<!doctype html>")
}

func TestResendVerification_MissingEmail(t *testing.T) {
	h, _ := newTestEmailAuthHandlers(t)

	e := echo.New()
	body := strings.NewReader(`{"email":""}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/resend-verification", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ResendVerification(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "email is required")
}

func TestResendVerification_NonexistentEmail(t *testing.T) {
	h, _ := newTestEmailAuthHandlers(t)

	e := echo.New()
	// Non-existent email should still return OK (don't reveal if email exists)
	body := strings.NewReader(`{"email":"nonexistent@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/resend-verification", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ResendVerification(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
