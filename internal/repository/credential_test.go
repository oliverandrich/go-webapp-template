// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository_test

import (
	"context"
	"testing"

	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"codeberg.org/oliverandrich/go-webapp-template/internal/repository"
	"codeberg.org/oliverandrich/go-webapp-template/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCreateCredential(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser", "Test User")

	cred := &models.Credential{
		UserID:       user.ID,
		CredentialID: []byte("test-cred-id"),
		PublicKey:    []byte("test-public-key"),
		AAGUID:       []byte("test-aaguid"),
		Name:         "My Passkey",
	}
	err := repo.CreateCredential(ctx, cred)

	require.NoError(t, err)
	assert.NotZero(t, cred.ID)
}

func TestGetCredentialsByUserID(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser", "Test User")
	testutil.NewTestCredential(t, db, user.ID, "cred-1")
	testutil.NewTestCredential(t, db, user.ID, "cred-2")

	creds, err := repo.GetCredentialsByUserID(ctx, user.ID)

	require.NoError(t, err)
	assert.Len(t, creds, 2)
}

func TestGetCredentialsByUserID_Empty(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser", "Test User")

	creds, err := repo.GetCredentialsByUserID(ctx, user.ID)

	require.NoError(t, err)
	assert.Empty(t, creds)
}

func TestUpdateCredentialSignCount(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser", "Test User")
	cred := testutil.NewTestCredential(t, db, user.ID, "my-cred")

	err := repo.UpdateCredentialSignCount(ctx, cred.CredentialID, 42)

	require.NoError(t, err)

	// Verify the update
	var updated models.Credential
	require.NoError(t, db.First(&updated, cred.ID).Error)
	assert.Equal(t, uint32(42), updated.SignCount)
}

func TestDeleteCredential(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser", "Test User")
	cred := testutil.NewTestCredential(t, db, user.ID, "to-delete")

	err := repo.DeleteCredential(ctx, cred.ID, user.ID)

	require.NoError(t, err)

	// Verify deletion
	var count int64
	db.Model(&models.Credential{}).Where("id = ?", cred.ID).Count(&count)
	assert.Zero(t, count)
}

func TestDeleteCredential_WrongUser(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user1 := testutil.NewTestUser(t, db, "user1", "User 1")
	user2 := testutil.NewTestUser(t, db, "user2", "User 2")
	cred := testutil.NewTestCredential(t, db, user1.ID, "user1-cred")

	// Try to delete user1's credential as user2
	err := repo.DeleteCredential(ctx, cred.ID, user2.ID)

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// Verify credential still exists
	var count int64
	db.Model(&models.Credential{}).Where("id = ?", cred.ID).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestDeleteCredential_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser", "Test User")

	err := repo.DeleteCredential(ctx, 999, user.ID)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCountUserCredentials(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser", "Test User")
	testutil.NewTestCredential(t, db, user.ID, "cred-1")
	testutil.NewTestCredential(t, db, user.ID, "cred-2")

	count, err := repo.CountUserCredentials(ctx, user.ID)

	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestCountUserCredentials_Zero(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser", "Test User")

	count, err := repo.CountUserCredentials(ctx, user.ID)

	require.NoError(t, err)
	assert.Zero(t, count)
}
