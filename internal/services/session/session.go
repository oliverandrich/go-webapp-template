// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package session

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/oliverandrich/go-webapp-template/internal/config"
)

// Data contains the session information stored in the cookie.
type Data struct { //nolint:govet // fieldalignment not critical
	UserID    int64     `json:"u"`
	Username  string    `json:"n"`
	ExpiresAt time.Time `json:"e"`
}

// Manager handles session cookie creation and parsing.
type Manager struct {
	sc         *securecookie.SecureCookie
	cookieName string
	maxAge     int
	secure     bool
}

// NewManager creates a new session manager.
func NewManager(cfg *config.SessionConfig, secure bool) (*Manager, error) {
	hashKey, err := resolveKey(cfg.HashKey, "hash")
	if err != nil {
		return nil, err
	}

	var blockKey []byte
	if cfg.BlockKey != "" {
		blockKey, err = hex.DecodeString(cfg.BlockKey)
		if err != nil {
			return nil, errors.New("invalid session block key: must be hex encoded")
		}
		if len(blockKey) != 32 {
			return nil, errors.New("invalid session block key: must be 32 bytes")
		}
	}

	sc := securecookie.New(hashKey, blockKey)
	sc.MaxAge(cfg.MaxAge)

	return &Manager{
		sc:         sc,
		cookieName: cfg.CookieName,
		maxAge:     cfg.MaxAge,
		secure:     secure,
	}, nil
}

// resolveKey resolves the key from config or generates one for development.
func resolveKey(keyHex, keyType string) ([]byte, error) {
	if keyHex != "" {
		key, err := hex.DecodeString(keyHex)
		if err != nil {
			return nil, errors.New("invalid session " + keyType + " key: must be hex encoded")
		}
		if len(key) != 32 {
			return nil, errors.New("invalid session " + keyType + " key: must be 32 bytes")
		}
		return key, nil
	}

	// Generate random key for development
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, errors.New("failed to generate session " + keyType + " key")
	}
	slog.Warn("No session "+keyType+" key configured, using random key (sessions will not persist across restarts)",
		"generated_key", hex.EncodeToString(key),
	)
	return key, nil
}

// Create creates a new session cookie for the given user.
func (m *Manager) Create(userID int64, username string) (*http.Cookie, error) {
	data := Data{
		UserID:    userID,
		Username:  username,
		ExpiresAt: time.Now().Add(time.Duration(m.maxAge) * time.Second),
	}

	encoded, err := m.sc.Encode(m.cookieName, data)
	if err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name:     m.cookieName,
		Value:    encoded,
		Path:     "/",
		MaxAge:   m.maxAge,
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

// Parse parses the session cookie from the request.
// Returns nil, nil if no valid session cookie is present.
func (m *Manager) Parse(r *http.Request) (*Data, error) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, nil
		}
		return nil, err
	}

	var data Data
	if err := m.sc.Decode(m.cookieName, cookie.Value, &data); err != nil {
		return nil, nil //nolint:nilerr // Invalid cookie is treated as no session
	}

	// Check expiration
	if time.Now().After(data.ExpiresAt) {
		return nil, nil
	}

	return &data, nil
}

// Clear returns a cookie that clears the session.
func (m *Manager) Clear() *http.Cookie {
	return &http.Cookie{
		Name:     m.cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// Flash cookie name.
const flashCookieName = "flash"

// FlashData contains temporary data that is cleared after reading.
type FlashData struct {
	RecoveryCodes []string `json:"rc,omitempty"`
}

// SetFlash creates a flash cookie with temporary data.
func (m *Manager) SetFlash(data *FlashData) (*http.Cookie, error) {
	encoded, err := m.sc.Encode(flashCookieName, data)
	if err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name:     flashCookieName,
		Value:    encoded,
		Path:     "/",
		MaxAge:   300, // 5 minutes max
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

// GetFlash reads and returns the flash data.
// Returns nil if no flash cookie exists.
func (m *Manager) GetFlash(r *http.Request) *FlashData {
	cookie, err := r.Cookie(flashCookieName)
	if err != nil {
		return nil
	}

	var data FlashData
	if err := m.sc.Decode(flashCookieName, cookie.Value, &data); err != nil {
		return nil
	}

	return &data
}

// ClearFlash returns a cookie that clears the flash data.
func (m *Manager) ClearFlash() *http.Cookie {
	return &http.Cookie{
		Name:     flashCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
	}
}
