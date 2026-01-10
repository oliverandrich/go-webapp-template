// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package models_test

import (
	"testing"

	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/stretchr/testify/assert"
)

func TestCredential_ToWebAuthn(t *testing.T) {
	cred := &models.Credential{
		CredentialID:    []byte("test-cred-id"),
		PublicKey:       []byte("test-public-key"),
		AAGUID:          []byte("test-aaguid"),
		SignCount:       42,
		BackupEligible:  true,
		BackupState:     true,
		AttestationType: "none",
	}

	webauthnCred := cred.ToWebAuthn()

	assert.Equal(t, []byte("test-cred-id"), webauthnCred.ID)
	assert.Equal(t, []byte("test-public-key"), webauthnCred.PublicKey)
	assert.Equal(t, []byte("test-aaguid"), webauthnCred.Authenticator.AAGUID)
	assert.Equal(t, uint32(42), webauthnCred.Authenticator.SignCount)
	assert.True(t, webauthnCred.Flags.BackupEligible)
	assert.True(t, webauthnCred.Flags.BackupState)
	assert.Equal(t, "none", webauthnCred.AttestationType)
}

func TestCredential_ToWebAuthn_WithTransports(t *testing.T) {
	cred := &models.Credential{
		CredentialID: []byte("cred"),
		PublicKey:    []byte("key"),
		Transports:   "usb,nfc,ble",
	}

	webauthnCred := cred.ToWebAuthn()

	assert.Len(t, webauthnCred.Transport, 3)
	assert.Contains(t, webauthnCred.Transport, protocol.AuthenticatorTransport("usb"))
	assert.Contains(t, webauthnCred.Transport, protocol.AuthenticatorTransport("nfc"))
	assert.Contains(t, webauthnCred.Transport, protocol.AuthenticatorTransport("ble"))
}

func TestCredential_ToWebAuthn_EmptyTransports(t *testing.T) {
	cred := &models.Credential{
		CredentialID: []byte("cred"),
		PublicKey:    []byte("key"),
		Transports:   "",
	}

	webauthnCred := cred.ToWebAuthn()

	assert.Empty(t, webauthnCred.Transport)
}

func TestTransportsFromWebAuthn(t *testing.T) {
	transports := []protocol.AuthenticatorTransport{
		protocol.USB,
		protocol.NFC,
		protocol.BLE,
	}

	result := models.TransportsFromWebAuthn(transports)

	assert.Equal(t, "usb,nfc,ble", result)
}

func TestTransportsFromWebAuthn_Empty(t *testing.T) {
	result := models.TransportsFromWebAuthn(nil)

	assert.Empty(t, result)
}

func TestTransportsFromWebAuthn_Single(t *testing.T) {
	transports := []protocol.AuthenticatorTransport{protocol.Internal}

	result := models.TransportsFromWebAuthn(transports)

	assert.Equal(t, "internal", result)
}

func TestAllModels(t *testing.T) {
	allModels := models.AllModels()

	assert.Len(t, allModels, 3) // User, Credential, RecoveryCode
}
