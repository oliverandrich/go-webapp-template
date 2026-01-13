// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository_test

import (
	"context"
	"testing"

	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"codeberg.org/oliverandrich/go-webapp-template/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCredential(t *testing.T) {
	_, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	user := testutil.NewTestUser(t, repo, "testuser")

	cred := &models.Credential{
		UserID:       user.ID,
		CredentialID: []byte("test-cred-id"),
		PublicKey:    []byte("test-public-key"),
		AAGUID:       []byte("test-aaguid"),
		Name:         "My Passkey",
	}
	err := repo.CreateCredential(ctx, cred)

	require.NoError(t, err)
}

func TestGetCredentialsByUserID(t *testing.T) {
	_, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	user := testutil.NewTestUser(t, repo, "testuser")
	testutil.NewTestCredential(t, repo, user.ID, "cred-1")
	testutil.NewTestCredential(t, repo, user.ID, "cred-2")

	creds, err := repo.GetCredentialsByUserID(ctx, user.ID)

	require.NoError(t, err)
	assert.Len(t, creds, 2)
}

func TestGetCredentialsByUserID_Empty(t *testing.T) {
	_, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	user := testutil.NewTestUser(t, repo, "testuser")

	creds, err := repo.GetCredentialsByUserID(ctx, user.ID)

	require.NoError(t, err)
	assert.Empty(t, creds)
}

func TestUpdateCredentialSignCount(t *testing.T) {
	db, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	user := testutil.NewTestUser(t, repo, "testuser")
	cred := testutil.NewTestCredential(t, repo, user.ID, "my-cred")

	err := repo.UpdateCredentialSignCount(ctx, cred.CredentialID, 42)

	require.NoError(t, err)

	// Verify the update
	var updated models.Credential
	require.NoError(t, db.GetContext(ctx, &updated, `SELECT * FROM credentials WHERE id = ?`, cred.ID))
	assert.Equal(t, uint32(42), updated.SignCount)
}

func TestDeleteCredential(t *testing.T) {
	db, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	user := testutil.NewTestUser(t, repo, "testuser")
	cred := testutil.NewTestCredential(t, repo, user.ID, "to-delete")

	err := repo.DeleteCredential(ctx, cred.ID, user.ID)

	require.NoError(t, err)

	// Verify deletion
	var count int64
	require.NoError(t, db.GetContext(ctx, &count, `SELECT COUNT(*) FROM credentials WHERE id = ?`, cred.ID))
	assert.Zero(t, count)
}

func TestDeleteCredential_WrongUser(t *testing.T) {
	db, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	user1 := testutil.NewTestUser(t, repo, "user1")
	user2 := testutil.NewTestUser(t, repo, "user2")
	cred := testutil.NewTestCredential(t, repo, user1.ID, "user1-cred")

	// Try to delete user1's credential as user2 - with sqlx this doesn't error, just affects 0 rows
	err := repo.DeleteCredential(ctx, cred.ID, user2.ID)

	require.NoError(t, err)

	// Verify credential still exists
	var count int64
	require.NoError(t, db.GetContext(ctx, &count, `SELECT COUNT(*) FROM credentials WHERE id = ?`, cred.ID))
	assert.Equal(t, int64(1), count)
}

func TestDeleteCredential_NotFound(t *testing.T) {
	_, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	user := testutil.NewTestUser(t, repo, "testuser")

	// With sqlx, deleting a non-existent row doesn't error
	err := repo.DeleteCredential(ctx, 999, user.ID)

	require.NoError(t, err)
}

func TestCountUserCredentials(t *testing.T) {
	_, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	user := testutil.NewTestUser(t, repo, "testuser")
	testutil.NewTestCredential(t, repo, user.ID, "cred-1")
	testutil.NewTestCredential(t, repo, user.ID, "cred-2")

	count, err := repo.CountUserCredentials(ctx, user.ID)

	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestCountUserCredentials_Zero(t *testing.T) {
	_, repo := testutil.NewTestDB(t)
	ctx := context.Background()

	user := testutil.NewTestUser(t, repo, "testuser")

	count, err := repo.CountUserCredentials(ctx, user.ID)

	require.NoError(t, err)
	assert.Zero(t, count)
}
