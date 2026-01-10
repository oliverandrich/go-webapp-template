// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package models_test

import (
	"encoding/binary"
	"testing"

	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestUser_WebAuthnID(t *testing.T) {
	user := &models.User{ID: 123}

	id := user.WebAuthnID()

	assert.Len(t, id, 8)
	assert.Equal(t, uint64(123), binary.BigEndian.Uint64(id))
}

func TestUser_WebAuthnID_LargeID(t *testing.T) {
	user := &models.User{ID: 9223372036854775807} // max int64

	id := user.WebAuthnID()

	assert.Len(t, id, 8)
	assert.Equal(t, uint64(9223372036854775807), binary.BigEndian.Uint64(id))
}

func TestUser_WebAuthnName(t *testing.T) {
	user := &models.User{Username: "testuser"}

	assert.Equal(t, "testuser", user.WebAuthnName())
}

func TestUser_WebAuthnDisplayName(t *testing.T) {
	user := &models.User{Username: "testuser"}

	assert.Equal(t, "testuser", user.WebAuthnDisplayName())
}

func TestUser_WebAuthnIcon(t *testing.T) {
	user := &models.User{}

	assert.Empty(t, user.WebAuthnIcon())
}

func TestUser_WebAuthnCredentials(t *testing.T) {
	user := &models.User{
		Credentials: []models.Credential{
			{
				CredentialID:    []byte("cred-1"),
				PublicKey:       []byte("key-1"),
				AAGUID:          []byte("aaguid-1"),
				SignCount:       5,
				Transports:      "usb,nfc",
				BackupEligible:  true,
				BackupState:     false,
				AttestationType: "none",
			},
			{
				CredentialID: []byte("cred-2"),
				PublicKey:    []byte("key-2"),
				AAGUID:       []byte("aaguid-2"),
				SignCount:    10,
			},
		},
	}

	creds := user.WebAuthnCredentials()

	assert.Len(t, creds, 2)
	assert.Equal(t, []byte("cred-1"), creds[0].ID)
	assert.Equal(t, []byte("key-1"), creds[0].PublicKey)
	assert.Equal(t, uint32(5), creds[0].Authenticator.SignCount)
}

func TestUser_WebAuthnCredentials_Empty(t *testing.T) {
	user := &models.User{}

	creds := user.WebAuthnCredentials()

	assert.Empty(t, creds)
}
