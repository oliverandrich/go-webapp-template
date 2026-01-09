// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository

import (
	"context"

	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"gorm.io/gorm"
)

// CreateCredential creates a new credential.
func (r *Repository) CreateCredential(ctx context.Context, cred *models.Credential) error {
	return r.db.WithContext(ctx).Create(cred).Error
}

// GetCredentialsByUserID retrieves all credentials for a user.
func (r *Repository) GetCredentialsByUserID(ctx context.Context, userID int64) ([]models.Credential, error) {
	var creds []models.Credential
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&creds).Error; err != nil {
		return nil, err
	}
	return creds, nil
}

// UpdateCredentialSignCount updates the sign count for a credential.
func (r *Repository) UpdateCredentialSignCount(ctx context.Context, credentialID []byte, signCount uint32) error {
	return r.db.WithContext(ctx).Model(&models.Credential{}).
		Where("credential_id = ?", credentialID).
		Update("sign_count", signCount).Error
}

// DeleteCredential deletes a credential by ID, ensuring it belongs to the given user.
func (r *Repository) DeleteCredential(ctx context.Context, id, userID int64) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&models.Credential{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CountUserCredentials counts the number of credentials for a user.
func (r *Repository) CountUserCredentials(ctx context.Context, userID int64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Credential{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
