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
	token := &models.EmailVerificationToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	return r.db.WithContext(ctx).Create(token).Error
}

// GetEmailVerificationToken retrieves a verification token by its hash.
// Returns ErrNotFound if the token doesn't exist.
func (r *Repository) GetEmailVerificationToken(ctx context.Context, tokenHash string) (*models.EmailVerificationToken, error) {
	var token models.EmailVerificationToken
	if err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteEmailVerificationToken deletes a verification token by ID.
func (r *Repository) DeleteEmailVerificationToken(ctx context.Context, tokenID int64) error {
	return r.db.WithContext(ctx).Delete(&models.EmailVerificationToken{}, tokenID).Error
}

// DeleteExpiredEmailVerificationTokens removes all expired tokens.
func (r *Repository) DeleteExpiredEmailVerificationTokens(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.EmailVerificationToken{}).Error
}

// DeleteUserEmailVerificationTokens deletes all verification tokens for a user.
func (r *Repository) DeleteUserEmailVerificationTokens(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.EmailVerificationToken{}).Error
}
