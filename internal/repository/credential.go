// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository

import (
	"context"

	"github.com/oliverandrich/go-webapp-template/internal/models"
)

// CreateCredential creates a new WebAuthn credential.
func (r *Repository) CreateCredential(ctx context.Context, cred *models.Credential) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO credentials (user_id, credential_id, public_key, attestation_type, transports, aaguid, sign_count, backup_eligible, backup_state, name)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cred.UserID, cred.CredentialID, cred.PublicKey, cred.AttestationType, cred.Transports,
		cred.AAGUID, cred.SignCount, cred.BackupEligible, cred.BackupState, cred.Name)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	cred.ID = id
	return nil
}

// GetCredentialsByUserID retrieves all credentials for a user.
func (r *Repository) GetCredentialsByUserID(ctx context.Context, userID int64) ([]models.Credential, error) {
	var creds []models.Credential
	err := r.db.SelectContext(ctx, &creds, `SELECT * FROM credentials WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	return creds, nil
}

// UpdateCredentialSignCount updates the sign count for a credential by credential_id bytes.
func (r *Repository) UpdateCredentialSignCount(ctx context.Context, credentialID []byte, signCount uint32) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE credentials SET sign_count = ? WHERE credential_id = ?`,
		signCount, credentialID)
	return err
}

// DeleteCredential deletes a credential.
func (r *Repository) DeleteCredential(ctx context.Context, credID, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM credentials WHERE id = ? AND user_id = ?`,
		credID, userID)
	return err
}

// CountUserCredentials counts the number of credentials for a user.
func (r *Repository) CountUserCredentials(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM credentials WHERE user_id = ?`, userID)
	return count, err
}
