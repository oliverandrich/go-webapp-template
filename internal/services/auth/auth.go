// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"

	"codeberg.org/oliverandrich/go-webapp-template/internal/config"
	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"codeberg.org/oliverandrich/go-webapp-template/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrRegistrationClosed = errors.New("registration is closed")
	ErrWeakPassword       = errors.New("password does not meet requirements")
	ErrInvalidEmail       = errors.New("invalid email format")
)

// dummyHash is used for constant-time login to prevent timing attacks
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)

type Service struct {
	repo              *repository.Repository
	config            *config.AuthConfig
	passwordValidator *PasswordValidator
}

func NewService(repo *repository.Repository, cfg *config.AuthConfig) *Service {
	return &Service{
		repo:              repo,
		config:            cfg,
		passwordValidator: DefaultPasswordValidator(),
	}
}

// PasswordValidator returns the password validator for use in handlers
func (s *Service) PasswordValidator() *PasswordValidator {
	return s.passwordValidator
}

// RegisterParams holds the parameters for user registration
type RegisterParams struct {
	Email    string
	Password string
	IsAdmin  bool
}

// ValidatePassword validates a password and returns the validation result
func (s *Service) ValidatePassword(password string, userAttributes ...string) ValidationResult {
	return s.passwordValidator.Validate(password, userAttributes...)
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, params RegisterParams) (*models.User, error) {
	// Validate email format
	if _, err := mail.ParseAddress(params.Email); err != nil {
		return nil, ErrInvalidEmail
	}

	// Check if registration is enabled
	if !s.config.IsRegistrationEnabled() {
		return nil, ErrRegistrationClosed
	}

	// Validate password
	validation := s.passwordValidator.Validate(params.Password, params.Email)
	if !validation.Valid {
		return nil, &PasswordValidationError{Errors: validation.Errors}
	}

	// Check if user already exists
	_, err := s.repo.GetUserByEmail(ctx, params.Email)
	if err == nil {
		return nil, ErrUserExists
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Email:        params.Email,
		PasswordHash: string(passwordHash),
		IsAdmin:      boolToInt64(params.IsAdmin),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	slog.Info("register_success", "user_id", user.ID, "email", params.Email)

	return user, nil
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// Login authenticates a user and returns the user if successful
func (s *Service) Login(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Constant-time: always perform bcrypt comparison to prevent timing attacks
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
			slog.Warn("login_failed", "email", email, "reason", "user_not_found")
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		slog.Warn("login_failed", "email", email, "reason", "invalid_password")
		return nil, ErrInvalidCredentials
	}

	slog.Info("login_success", "user_id", user.ID, "email", email)
	return user, nil
}

// ChangePassword changes a user's password (when they know their current password)
func (s *Service) ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify current password
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Validate new password
	validation := s.passwordValidator.Validate(newPassword, user.Email)
	if !validation.Valid {
		return &PasswordValidationError{Errors: validation.Errors}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.repo.UpdateUserPassword(ctx, userID, string(passwordHash)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// SetAdmin sets or removes admin status for a user
func (s *Service) SetAdmin(ctx context.Context, userID int64, isAdmin bool) error {
	return s.repo.SetUserAdmin(ctx, userID, isAdmin)
}

// EnsureAdmin ensures at least one admin exists, creating one if needed
func (s *Service) EnsureAdmin(ctx context.Context, email, password string) error {
	count, err := s.repo.CountAdmins(ctx)
	if err != nil {
		return fmt.Errorf("failed to count admins: %w", err)
	}

	if count > 0 {
		return nil // Admin already exists
	}

	// Create admin user
	user, err := s.Register(ctx, RegisterParams{
		Email:    email,
		Password: password,
		IsAdmin:  true,
	})
	if err != nil && !errors.Is(err, ErrUserExists) {
		return fmt.Errorf("failed to create admin: %w", err)
	}

	// If user exists, make them admin
	if errors.Is(err, ErrUserExists) {
		existingUser, err := s.repo.GetUserByEmail(ctx, email)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		if err := s.SetAdmin(ctx, existingUser.ID, true); err != nil {
			return fmt.Errorf("failed to set admin: %w", err)
		}
	} else {
		// Set newly created user as admin
		if err := s.SetAdmin(ctx, user.ID, true); err != nil {
			return fmt.Errorf("failed to set admin: %w", err)
		}
	}

	return nil
}
