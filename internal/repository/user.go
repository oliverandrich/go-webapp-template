// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository

import (
	"context"

	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
)

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
