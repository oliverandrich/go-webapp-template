// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
)

// ErrNotFound is returned when a record is not found
var ErrNotFound = errors.New("record not found")

// Repository wraps GORM for database operations
type Repository struct {
	db *gorm.DB
}

// New creates a new Repository instance
func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// DB returns the underlying GORM DB for direct access
func (r *Repository) DB() *gorm.DB {
	return r.db
}

// wrapError converts GORM errors to repository errors
func wrapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

// ===== User Methods =====

// GetUserByID retrieves a user by their ID
func (r *Repository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, wrapError(err)
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by their email address
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, wrapError(err)
	}
	return &user, nil
}

// CreateUser creates a new user in the database
func (r *Repository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// UpdateUser updates an existing user in the database
func (r *Repository) UpdateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// DeleteUser deletes a user by their ID
func (r *Repository) DeleteUser(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

// ListUsers returns all users ordered by creation date (newest first)
func (r *Repository) ListUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// CountUsers returns the total number of users
func (r *Repository) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountAdmins returns the number of admin users
func (r *Repository) CountAdmins(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("is_admin = 1").Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// UpdateUserPassword updates a user's password
func (r *Repository) UpdateUserPassword(ctx context.Context, id int64, passwordHash string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("password_hash", passwordHash).Error
}

// SetUserAdmin sets or removes admin status for a user
func (r *Repository) SetUserAdmin(ctx context.Context, id int64, isAdmin bool) error {
	val := int64(0)
	if isAdmin {
		val = 1
	}
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("is_admin", val).Error
}
