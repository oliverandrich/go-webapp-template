// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package models

import (
	"encoding/binary"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

// User represents an authenticated user with WebAuthn credentials.
type User struct { //nolint:govet // fieldalignment not critical for models
	ID              int64        `gorm:"primaryKey" json:"id"`
	Username        string       `gorm:"uniqueIndex;not null;size:64" json:"username"`
	DisplayName     string       `gorm:"not null;size:128" json:"display_name"`
	Email           *string      `gorm:"uniqueIndex;size:255" json:"email,omitempty"`
	EmailVerified   bool         `gorm:"not null;default:false" json:"email_verified"`
	EmailVerifiedAt *time.Time   `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
	Credentials     []Credential `gorm:"foreignKey:UserID" json:"credentials,omitempty"`
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

// WebAuthnDisplayName returns the user's display name.
func (u *User) WebAuthnDisplayName() string {
	return u.DisplayName
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
