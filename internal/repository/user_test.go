// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/oliverandrich/go-webapp-template/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	_, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	user, err := repo.CreateUser(ctx, "testuser")

	require.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.NotZero(t, user.CreatedAt)
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	_, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	_, err := repo.CreateUser(ctx, "testuser")
	require.NoError(t, err)

	_, err = repo.CreateUser(ctx, "testuser")

	assert.Error(t, err)
}

func TestGetUserByID(t *testing.T) {
	_, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	created, err := repo.CreateUser(ctx, "testuser")
	require.NoError(t, err)

	retrieved, err := repo.GetUserByID(ctx, created.ID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Username, retrieved.Username)
}

func TestGetUserByID_NotFound(t *testing.T) {
	_, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	_, err := repo.GetUserByID(ctx, 999)

	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestGetUserByID_WithCredentials(t *testing.T) {
	_, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	user := testutil.NewTestUser(t, repo, "testuser")
	testutil.NewTestCredential(t, repo, user.ID, "credential-1")
	testutil.NewTestCredential(t, repo, user.ID, "credential-2")

	// With sqlx we need to load credentials separately
	creds, err := repo.GetCredentialsByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, creds, 2)
}

func TestGetUserByUsername(t *testing.T) {
	_, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	created, err := repo.CreateUser(ctx, "testuser")
	require.NoError(t, err)

	retrieved, err := repo.GetUserByUsername(ctx, "testuser")

	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
}

func TestGetUserByUsername_NotFound(t *testing.T) {
	_, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	_, err := repo.GetUserByUsername(ctx, "nonexistent")

	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestUserExists(t *testing.T) {
	_, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	_, err := repo.CreateUser(ctx, "testuser")
	require.NoError(t, err)

	exists, err := repo.UserExists(ctx, "testuser")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestUserExists_NotFound(t *testing.T) {
	_, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	exists, err := repo.UserExists(ctx, "nonexistent")

	require.NoError(t, err)
	assert.False(t, exists)
}
