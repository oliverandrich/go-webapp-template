// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository

import (
	"context"

	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// CreateRecoveryCodes creates recovery codes for a user.
func (r *Repository) CreateRecoveryCodes(ctx context.Context, userID int64, codeHashes []string) error {
	for _, hash := range codeHashes {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO recovery_codes (user_id, code_hash) VALUES (?, ?)`,
			userID, hash)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetUnusedRecoveryCodes retrieves unused recovery codes for a user.
func (r *Repository) GetUnusedRecoveryCodes(ctx context.Context, userID int64) ([]models.RecoveryCode, error) {
	var codes []models.RecoveryCode
	err := r.db.SelectContext(ctx, &codes, `SELECT * FROM recovery_codes WHERE user_id = ? AND used = 0`, userID)
	if err != nil {
		return nil, err
	}
	return codes, nil
}

// GetUnusedRecoveryCodeCount returns the count of unused recovery codes.
func (r *Repository) GetUnusedRecoveryCodeCount(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM recovery_codes WHERE user_id = ? AND used = 0`, userID)
	return count, err
}

// MarkRecoveryCodeUsed marks a recovery code as used.
func (r *Repository) MarkRecoveryCodeUsed(ctx context.Context, codeID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE recovery_codes SET used = 1, used_at = CURRENT_TIMESTAMP WHERE id = ?`,
		codeID)
	return err
}

// DeleteRecoveryCodes deletes all recovery codes for a user.
func (r *Repository) DeleteRecoveryCodes(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM recovery_codes WHERE user_id = ?`, userID)
	return err
}

// HasRecoveryCodes checks if a user has any recovery codes.
func (r *Repository) HasRecoveryCodes(ctx context.Context, userID int64) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM recovery_codes WHERE user_id = ?)`, userID)
	return exists, err
}

// ValidateAndUseRecoveryCode validates and marks a recovery code as used.
func (r *Repository) ValidateAndUseRecoveryCode(ctx context.Context, userID int64, code string) (bool, error) {
	codes, err := r.GetUnusedRecoveryCodes(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, c := range codes {
		if bcrypt.CompareHashAndPassword([]byte(c.CodeHash), []byte(code)) == nil {
			if markErr := r.MarkRecoveryCodeUsed(ctx, c.ID); markErr != nil {
				return false, markErr
			}
			return true, nil
		}
	}

	return false, nil
}
