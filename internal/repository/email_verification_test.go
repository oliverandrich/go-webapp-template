// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository_test

import (
	"context"
	"testing"
	"time"

	"codeberg.org/oliverandrich/go-webapp-template/internal/repository"
	"codeberg.org/oliverandrich/go-webapp-template/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCreateEmailVerificationToken(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser")
	tokenHash := "abc123hash"
	expiresAt := time.Now().Add(24 * time.Hour)

	err := repo.CreateEmailVerificationToken(ctx, user.ID, tokenHash, expiresAt)

	require.NoError(t, err)

	// Verify token was created
	token, err := repo.GetEmailVerificationToken(ctx, tokenHash)
	require.NoError(t, err)
	assert.Equal(t, user.ID, token.UserID)
	assert.Equal(t, tokenHash, token.TokenHash)
	assert.WithinDuration(t, expiresAt, token.ExpiresAt, time.Second)
}

func TestGetEmailVerificationToken(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser")
	tokenHash := "abc123hash"
	expiresAt := time.Now().Add(24 * time.Hour)

	err := repo.CreateEmailVerificationToken(ctx, user.ID, tokenHash, expiresAt)
	require.NoError(t, err)

	token, err := repo.GetEmailVerificationToken(ctx, tokenHash)

	require.NoError(t, err)
	assert.NotZero(t, token.ID)
	assert.Equal(t, user.ID, token.UserID)
	assert.Equal(t, tokenHash, token.TokenHash)
}

func TestGetEmailVerificationToken_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	_, err := repo.GetEmailVerificationToken(ctx, "nonexistent")

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestDeleteEmailVerificationToken(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser")
	tokenHash := "abc123hash"
	expiresAt := time.Now().Add(24 * time.Hour)

	err := repo.CreateEmailVerificationToken(ctx, user.ID, tokenHash, expiresAt)
	require.NoError(t, err)

	token, err := repo.GetEmailVerificationToken(ctx, tokenHash)
	require.NoError(t, err)

	err = repo.DeleteEmailVerificationToken(ctx, token.ID)
	require.NoError(t, err)

	// Should not be found anymore
	_, err = repo.GetEmailVerificationToken(ctx, tokenHash)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestDeleteUserEmailVerificationTokens(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser")
	expiresAt := time.Now().Add(24 * time.Hour)

	// Create multiple tokens for the user
	err := repo.CreateEmailVerificationToken(ctx, user.ID, "token1", expiresAt)
	require.NoError(t, err)
	err = repo.CreateEmailVerificationToken(ctx, user.ID, "token2", expiresAt)
	require.NoError(t, err)

	err = repo.DeleteUserEmailVerificationTokens(ctx, user.ID)
	require.NoError(t, err)

	// Both should be deleted
	_, err = repo.GetEmailVerificationToken(ctx, "token1")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	_, err = repo.GetEmailVerificationToken(ctx, "token2")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestDeleteExpiredEmailVerificationTokens(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser")

	// Create an expired token
	expiredAt := time.Now().Add(-1 * time.Hour)
	err := repo.CreateEmailVerificationToken(ctx, user.ID, "expired", expiredAt)
	require.NoError(t, err)

	// Create a valid token
	validAt := time.Now().Add(24 * time.Hour)
	err = repo.CreateEmailVerificationToken(ctx, user.ID, "valid", validAt)
	require.NoError(t, err)

	err = repo.DeleteExpiredEmailVerificationTokens(ctx)
	require.NoError(t, err)

	// Expired should be deleted
	_, err = repo.GetEmailVerificationToken(ctx, "expired")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// Valid should still exist
	token, err := repo.GetEmailVerificationToken(ctx, "valid")
	require.NoError(t, err)
	assert.Equal(t, "valid", token.TokenHash)
}

// Tests for email-related user methods

func TestCreateUserWithEmail(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user, err := repo.CreateUserWithEmail(ctx, "test@example.com")

	require.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.Equal(t, "test@example.com", user.Username) // Username = email
	require.NotNil(t, user.Email)
	assert.Equal(t, "test@example.com", *user.Email)
	assert.False(t, user.EmailVerified)
}

func TestCreateUserWithEmail_DuplicateEmail(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	_, err := repo.CreateUserWithEmail(ctx, "test@example.com")
	require.NoError(t, err)

	_, err = repo.CreateUserWithEmail(ctx, "test@example.com")

	assert.Error(t, err)
}

func TestGetUserByEmail(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	created, err := repo.CreateUserWithEmail(ctx, "test@example.com")
	require.NoError(t, err)

	retrieved, err := repo.GetUserByEmail(ctx, "test@example.com")

	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	require.NotNil(t, retrieved.Email)
	assert.Equal(t, "test@example.com", *retrieved.Email)
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	_, err := repo.GetUserByEmail(ctx, "nonexistent@example.com")

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestEmailExists(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	_, err := repo.CreateUserWithEmail(ctx, "test@example.com")
	require.NoError(t, err)

	exists, err := repo.EmailExists(ctx, "test@example.com")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestEmailExists_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	exists, err := repo.EmailExists(ctx, "nonexistent@example.com")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMarkEmailVerified(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user, err := repo.CreateUserWithEmail(ctx, "test@example.com")
	require.NoError(t, err)
	assert.False(t, user.EmailVerified)

	err = repo.MarkEmailVerified(ctx, user.ID)
	require.NoError(t, err)

	// Retrieve and verify
	updated, err := repo.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
	assert.True(t, updated.EmailVerified)
	assert.NotNil(t, updated.EmailVerifiedAt)
}
