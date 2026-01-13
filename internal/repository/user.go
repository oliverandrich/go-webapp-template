// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository

import (
	"context"
	"database/sql"
	"errors"

	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
)

// CreateUser creates a new user with only a username.
func (r *Repository) CreateUser(ctx context.Context, username string) (*models.User, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO users (username) VALUES (?)`,
		username)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetUserByID(ctx, id)
}

// CreateUserWithEmail creates a new user with email.
func (r *Repository) CreateUserWithEmail(ctx context.Context, email string) (*models.User, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO users (username, email) VALUES (?, ?)`,
		email, email)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetUserByID(ctx, id)
}

// GetUserByID retrieves a user by ID.
func (r *Repository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username.
func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE username = ?`, username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email.
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE email = ?`, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UserExists checks if a user with the given username exists.
func (r *Repository) UserExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)`, username)
	return exists, err
}

// EmailExists checks if a user with the given email exists.
func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return exists, err
}

// MarkEmailVerified marks a user's email as verified.
func (r *Repository) MarkEmailVerified(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET email_verified = 1, email_verified_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		userID)
	return err
}
