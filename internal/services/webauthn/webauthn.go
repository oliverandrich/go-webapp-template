// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package webauthn

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/oliverandrich/go-webapp-template/internal/config"
)

const sessionTTL = 2 * time.Minute

// Service provides WebAuthn functionality.
type Service struct {
	wa       *webauthn.WebAuthn
	sessions *sessionStore
}

// NewService creates a new WebAuthn service.
func NewService(cfg *config.WebAuthnConfig) (*Service, error) {
	wconfig := &webauthn.Config{
		RPDisplayName: cfg.RPDisplayName,
		RPID:          cfg.RPID,
		RPOrigins:     []string{cfg.RPOrigin},
	}

	wa, err := webauthn.New(wconfig)
	if err != nil {
		return nil, err
	}

	return &Service{
		wa:       wa,
		sessions: newSessionStore(),
	}, nil
}

// WebAuthn returns the underlying webauthn.WebAuthn instance.
func (s *Service) WebAuthn() *webauthn.WebAuthn {
	return s.wa
}

// StoreRegistrationSession stores a registration session for a user.
func (s *Service) StoreRegistrationSession(userID int64, data *webauthn.SessionData) {
	s.sessions.store(registrationKey(userID), data)
}

// GetRegistrationSession retrieves and removes a registration session.
func (s *Service) GetRegistrationSession(userID int64) (*webauthn.SessionData, error) {
	return s.sessions.get(registrationKey(userID))
}

// StoreLoginSession stores a login session for a user.
func (s *Service) StoreLoginSession(userID int64, data *webauthn.SessionData) {
	s.sessions.store(loginKey(userID), data)
}

// GetLoginSession retrieves and removes a login session.
func (s *Service) GetLoginSession(userID int64) (*webauthn.SessionData, error) {
	return s.sessions.get(loginKey(userID))
}

// StoreDiscoverableSession stores a discoverable login session (usernameless).
func (s *Service) StoreDiscoverableSession(sessionID string, data *webauthn.SessionData) {
	s.sessions.store("discoverable:"+sessionID, data)
}

// GetDiscoverableSession retrieves and removes a discoverable login session.
func (s *Service) GetDiscoverableSession(sessionID string) (*webauthn.SessionData, error) {
	return s.sessions.get("discoverable:" + sessionID)
}

func registrationKey(userID int64) string {
	return fmt.Sprintf("registration:%d", userID)
}

func loginKey(userID int64) string {
	return fmt.Sprintf("login:%d", userID)
}

// sessionStore provides thread-safe session storage with TTL.
type sessionStore struct { //nolint:govet // fieldalignment not critical
	mu       sync.RWMutex
	sessions map[string]*sessionEntry
}

type sessionEntry struct {
	data      *webauthn.SessionData
	expiresAt time.Time
}

func newSessionStore() *sessionStore {
	ss := &sessionStore{
		sessions: make(map[string]*sessionEntry),
	}
	go ss.cleanup()
	return ss
}

func (s *sessionStore) store(key string, data *webauthn.SessionData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[key] = &sessionEntry{
		data:      data,
		expiresAt: time.Now().Add(sessionTTL),
	}
}

func (s *sessionStore) get(key string) (*webauthn.SessionData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.sessions[key]
	if !ok {
		return nil, errors.New("session not found")
	}

	delete(s.sessions, key)

	if time.Now().After(entry.expiresAt) {
		return nil, errors.New("session expired")
	}

	return entry.data, nil
}

func (s *sessionStore) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, entry := range s.sessions {
			if now.After(entry.expiresAt) {
				delete(s.sessions, key)
			}
		}
		s.mu.Unlock()
	}
}
