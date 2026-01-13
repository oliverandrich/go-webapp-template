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
type Credential struct { //nolint:govet // fieldalignment: readability over optimization
	ID              int64     `db:"id" json:"id"`
	UserID          int64     `db:"user_id" json:"user_id"`
	CredentialID    []byte    `db:"credential_id" json:"-"`
	PublicKey       []byte    `db:"public_key" json:"-"`
	AAGUID          []byte    `db:"aaguid" json:"-"`
	SignCount       uint32    `db:"sign_count" json:"-"`
	Transports      string    `db:"transports" json:"-"` // comma-separated
	Name            string    `db:"name" json:"name"`
	BackupEligible  bool      `db:"backup_eligible" json:"-"`
	BackupState     bool      `db:"backup_state" json:"-"`
	AttestationType string    `db:"attestation_type" json:"-"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
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
