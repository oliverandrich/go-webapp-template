// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository

import (
	"context"
	"time"

	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
)

// CreateEmailVerificationToken creates a new email verification token.
func (r *Repository) CreateEmailVerificationToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO email_verification_tokens (user_id, token_hash, expires_at) VALUES (?, ?, ?)`,
		userID, tokenHash, expiresAt)
	return err
}

// GetEmailVerificationToken retrieves an email verification token by hash.
func (r *Repository) GetEmailVerificationToken(ctx context.Context, tokenHash string) (*models.EmailVerificationToken, error) {
	var token models.EmailVerificationToken
	err := r.db.GetContext(ctx, &token, `SELECT * FROM email_verification_tokens WHERE token_hash = ?`, tokenHash)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteEmailVerificationToken deletes a token by ID.
func (r *Repository) DeleteEmailVerificationToken(ctx context.Context, tokenID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM email_verification_tokens WHERE id = ?`, tokenID)
	return err
}

// DeleteUserEmailVerificationTokens deletes all tokens for a user.
func (r *Repository) DeleteUserEmailVerificationTokens(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM email_verification_tokens WHERE user_id = ?`, userID)
	return err
}

// DeleteExpiredEmailVerificationTokens deletes expired tokens.
func (r *Repository) DeleteExpiredEmailVerificationTokens(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM email_verification_tokens WHERE expires_at < ?`, time.Now())
	return err
}
