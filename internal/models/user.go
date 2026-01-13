// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package models

import (
	"encoding/binary"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

// User represents an authenticated user with WebAuthn credentials.
type User struct { //nolint:govet // fieldalignment: readability over optimization
	ID              int64        `db:"id" json:"id"`
	Username        string       `db:"username" json:"username"`
	Email           *string      `db:"email" json:"email,omitempty"`
	EmailVerified   bool         `db:"email_verified" json:"email_verified"`
	EmailVerifiedAt *time.Time   `db:"email_verified_at" json:"email_verified_at,omitempty"`
	CreatedAt       time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time    `db:"updated_at" json:"updated_at"`
	Credentials     []Credential `db:"-" json:"credentials,omitempty"`
}

// WebAuthnID returns the user's ID as a byte slice for WebAuthn.
func (u *User) WebAuthnID() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(u.ID)) //nolint:gosec // user IDs are always positive
	return buf
}

// WebAuthnName returns the user's username.
func (u *User) WebAuthnName() string {
	return u.Username
}

// WebAuthnDisplayName returns the user's display name (username).
func (u *User) WebAuthnDisplayName() string {
	return u.Username
}

// WebAuthnCredentials returns the user's WebAuthn credentials.
func (u *User) WebAuthnCredentials() []webauthn.Credential {
	creds := make([]webauthn.Credential, len(u.Credentials))
	for i, c := range u.Credentials {
		creds[i] = c.ToWebAuthn()
	}
	return creds
}

// WebAuthnIcon returns an empty string (deprecated in WebAuthn spec).
func (u *User) WebAuthnIcon() string {
	return ""
}
