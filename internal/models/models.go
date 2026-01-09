// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package models

import (
	"encoding/binary"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// Example is a sample model to demonstrate GORM usage.
// Replace or extend with your own models.
type Example struct { //nolint:govet // fieldalignment not critical for models
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// User represents an authenticated user with WebAuthn credentials.
type User struct { //nolint:govet // fieldalignment not critical for models
	ID          int64        `gorm:"primaryKey" json:"id"`
	Username    string       `gorm:"uniqueIndex;not null;size:64" json:"username"`
	DisplayName string       `gorm:"not null;size:128" json:"display_name"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Credentials []Credential `gorm:"foreignKey:UserID" json:"credentials,omitempty"`
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

// Credential stores a WebAuthn credential for a user.
type Credential struct { //nolint:govet // fieldalignment not critical for models
	ID              int64     `gorm:"primaryKey" json:"id"`
	UserID          int64     `gorm:"not null;index" json:"user_id"`
	CredentialID    []byte    `gorm:"not null;uniqueIndex" json:"-"`
	PublicKey       []byte    `gorm:"not null" json:"-"`
	AAGUID          []byte    `gorm:"size:16" json:"-"`
	SignCount       uint32    `gorm:"not null;default:0" json:"-"`
	Transports      string    `gorm:"size:256" json:"-"` // comma-separated
	Name            string    `gorm:"size:128" json:"name"`
	BackupEligible  bool      `gorm:"not null;default:false" json:"-"`
	BackupState     bool      `gorm:"not null;default:false" json:"-"`
	AttestationType string    `gorm:"size:64" json:"-"`
	CreatedAt       time.Time `json:"created_at"`
}

// ToWebAuthn converts the database credential to a webauthn.Credential.
func (c *Credential) ToWebAuthn() webauthn.Credential {
	var transports []protocol.AuthenticatorTransport
	if c.Transports != "" {
		for t := range strings.SplitSeq(c.Transports, ",") {
			transports = append(transports, protocol.AuthenticatorTransport(t))
		}
	}
	return webauthn.Credential{
		ID:              c.CredentialID,
		PublicKey:       c.PublicKey,
		AttestationType: c.AttestationType,
		Transport:       transports,
		Flags: webauthn.CredentialFlags{
			UserPresent:    true,
			UserVerified:   true,
			BackupEligible: c.BackupEligible,
			BackupState:    c.BackupState,
		},
		Authenticator: webauthn.Authenticator{
			AAGUID:       c.AAGUID,
			SignCount:    c.SignCount,
			CloneWarning: false,
		},
	}
}

// TransportsFromWebAuthn converts WebAuthn transports to a comma-separated string.
func TransportsFromWebAuthn(transports []protocol.AuthenticatorTransport) string {
	strs := make([]string, len(transports))
	for i, t := range transports {
		strs[i] = string(t)
	}
	return strings.Join(strs, ",")
}

// AllModels returns all models for database migration.
func AllModels() []any {
	return []any{
		&Example{},
		&User{},
		&Credential{},
	}
}
