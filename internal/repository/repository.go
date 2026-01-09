// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository

import (
	"context"

	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"gorm.io/gorm"
)

// Repository provides data access methods.
type Repository struct {
	db *gorm.DB
}

// New creates a new Repository.
func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// DB returns the underlying database connection.
func (r *Repository) DB() *gorm.DB {
	return r.db
}

// Example methods - replace with your own

// CreateExample creates a new example record.
func (r *Repository) CreateExample(ctx context.Context, name string) (*models.Example, error) {
	example := &models.Example{Name: name}
	if err := r.db.WithContext(ctx).Create(example).Error; err != nil {
		return nil, err
	}
	return example, nil
}

// GetExample retrieves an example by ID.
func (r *Repository) GetExample(ctx context.Context, id int64) (*models.Example, error) {
	var example models.Example
	if err := r.db.WithContext(ctx).First(&example, id).Error; err != nil {
		return nil, err
	}
	return &example, nil
}

// ListExamples returns all examples.
func (r *Repository) ListExamples(ctx context.Context) ([]models.Example, error) {
	var examples []models.Example
	if err := r.db.WithContext(ctx).Find(&examples).Error; err != nil {
		return nil, err
	}
	return examples, nil
}

// DeleteExample deletes an example by ID.
func (r *Repository) DeleteExample(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Example{}, id).Error
}

// User methods

// CreateUser creates a new user.
func (r *Repository) CreateUser(ctx context.Context, username, displayName string) (*models.User, error) {
	user := &models.User{
		Username:    username,
		DisplayName: displayName,
	}
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByID retrieves a user by ID with preloaded credentials.
func (r *Repository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Credentials").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username with preloaded credentials.
func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Credentials").Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UserExists checks if a user with the given username exists.
func (r *Repository) UserExists(ctx context.Context, username string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Credential methods

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
