// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package models

import (
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

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
