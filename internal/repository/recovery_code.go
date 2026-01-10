// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository

import (
	"context"
	"time"

	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// CreateRecoveryCodes creates new recovery codes for a user.
func (r *Repository) CreateRecoveryCodes(ctx context.Context, userID int64, codeHashes []string) error {
	codes := make([]models.RecoveryCode, len(codeHashes))
	for i, hash := range codeHashes {
		codes[i] = models.RecoveryCode{
			UserID:   userID,
			CodeHash: hash,
		}
	}
	return r.db.WithContext(ctx).Create(&codes).Error
}

// GetUnusedRecoveryCodeCount returns the number of unused recovery codes for a user.
func (r *Repository) GetUnusedRecoveryCodeCount(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.RecoveryCode{}).
		Where("user_id = ? AND used = ?", userID, false).
		Count(&count).Error
	return count, err
}

// ValidateAndUseRecoveryCode checks if the code is valid and marks it as used.
func (r *Repository) ValidateAndUseRecoveryCode(ctx context.Context, userID int64, code string) (bool, error) {
	var codes []models.RecoveryCode
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND used = ?", userID, false).
		Find(&codes).Error; err != nil {
		return false, err
	}

	for _, rc := range codes {
		if bcrypt.CompareHashAndPassword([]byte(rc.CodeHash), []byte(code)) == nil {
			now := time.Now()
			rc.Used = true
			rc.UsedAt = &now
			if err := r.db.WithContext(ctx).Save(&rc).Error; err != nil {
				return false, err
			}
			return true, nil
		}
	}

	return false, nil
}

// DeleteRecoveryCodes deletes all recovery codes for a user.
func (r *Repository) DeleteRecoveryCodes(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.RecoveryCode{}).Error
}

// HasRecoveryCodes checks if the user has any recovery codes.
func (r *Repository) HasRecoveryCodes(ctx context.Context, userID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.RecoveryCode{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count > 0, err
}
