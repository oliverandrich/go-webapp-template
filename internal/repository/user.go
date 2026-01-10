// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository

import (
	"context"

	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
)

// CreateUser creates a new user.
func (r *Repository) CreateUser(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{
		Username: username,
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

// CreateUserWithEmail creates a new user with email as the primary identifier.
// Username is set to the email address.
func (r *Repository) CreateUserWithEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{
		Username: email, // Use email as username for WebAuthn compatibility
		Email:    &email,
	}
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email with preloaded credentials.
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Credentials").Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// EmailExists checks if a user with the given email exists.
func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// MarkEmailVerified marks a user's email as verified.
func (r *Repository) MarkEmailVerified(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"email_verified":    true,
			"email_verified_at": r.db.NowFunc(),
		}).Error
}
