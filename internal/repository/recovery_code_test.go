// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository_test

import (
	"context"
	"testing"

	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"codeberg.org/oliverandrich/go-webapp-template/internal/repository"
	"codeberg.org/oliverandrich/go-webapp-template/internal/services/recovery"
	"codeberg.org/oliverandrich/go-webapp-template/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRecoveryCodes(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser", "Test User")

	svc := recovery.NewService()
	_, hashes, err := svc.GenerateCodes(8)
	require.NoError(t, err)

	err = repo.CreateRecoveryCodes(ctx, user.ID, hashes)

	require.NoError(t, err)

	// Verify codes were created
	var count int64
	db.Model(&models.RecoveryCode{}).Where("user_id = ?", user.ID).Count(&count)
	assert.Equal(t, int64(8), count)
}

func TestGetUnusedRecoveryCodeCount(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser", "Test User")

	svc := recovery.NewService()
	_, hashes, err := svc.GenerateCodes(8)
	require.NoError(t, err)
	require.NoError(t, repo.CreateRecoveryCodes(ctx, user.ID, hashes))

	count, err := repo.GetUnusedRecoveryCodeCount(ctx, user.ID)

	require.NoError(t, err)
	assert.Equal(t, int64(8), count)
}

func TestGetUnusedRecoveryCodeCount_NoUser(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	count, err := repo.GetUnusedRecoveryCodeCount(ctx, 999)

	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestValidateAndUseRecoveryCode(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser", "Test User")

	svc := recovery.NewService()
	plaintexts, hashes, err := svc.GenerateCodes(3)
	require.NoError(t, err)
	require.NoError(t, repo.CreateRecoveryCodes(ctx, user.ID, hashes))

	// Validate a correct code (normalized)
	normalized := recovery.NormalizeCode(plaintexts[0])
	valid, err := repo.ValidateAndUseRecoveryCode(ctx, user.ID, normalized)

	require.NoError(t, err)
	assert.True(t, valid)

	// Check count decreased
	count, err := repo.GetUnusedRecoveryCodeCount(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestValidateAndUseRecoveryCode_InvalidCode(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser", "Test User")

	svc := recovery.NewService()
	_, hashes, err := svc.GenerateCodes(3)
	require.NoError(t, err)
	require.NoError(t, repo.CreateRecoveryCodes(ctx, user.ID, hashes))

	valid, err := repo.ValidateAndUseRecoveryCode(ctx, user.ID, "invalidcode12")

	require.NoError(t, err)
	assert.False(t, valid)
}

func TestValidateAndUseRecoveryCode_AlreadyUsed(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser", "Test User")

	svc := recovery.NewService()
	plaintexts, hashes, err := svc.GenerateCodes(3)
	require.NoError(t, err)
	require.NoError(t, repo.CreateRecoveryCodes(ctx, user.ID, hashes))

	normalized := recovery.NormalizeCode(plaintexts[0])

	// First use
	valid, err := repo.ValidateAndUseRecoveryCode(ctx, user.ID, normalized)
	require.NoError(t, err)
	assert.True(t, valid)

	// Second use (should fail)
	valid, err = repo.ValidateAndUseRecoveryCode(ctx, user.ID, normalized)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestValidateAndUseRecoveryCode_WrongUser(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user1 := testutil.NewTestUser(t, db, "user1", "User 1")
	user2 := testutil.NewTestUser(t, db, "user2", "User 2")

	svc := recovery.NewService()
	plaintexts, hashes, err := svc.GenerateCodes(3)
	require.NoError(t, err)
	require.NoError(t, repo.CreateRecoveryCodes(ctx, user1.ID, hashes))

	// Try to use user1's code as user2
	normalized := recovery.NormalizeCode(plaintexts[0])
	valid, err := repo.ValidateAndUseRecoveryCode(ctx, user2.ID, normalized)

	require.NoError(t, err)
	assert.False(t, valid)
}

func TestDeleteRecoveryCodes(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser", "Test User")

	svc := recovery.NewService()
	_, hashes, err := svc.GenerateCodes(8)
	require.NoError(t, err)
	require.NoError(t, repo.CreateRecoveryCodes(ctx, user.ID, hashes))

	err = repo.DeleteRecoveryCodes(ctx, user.ID)

	require.NoError(t, err)

	count, err := repo.GetUnusedRecoveryCodeCount(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestDeleteRecoveryCodes_OnlyAffectsUser(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user1 := testutil.NewTestUser(t, db, "user1", "User 1")
	user2 := testutil.NewTestUser(t, db, "user2", "User 2")

	svc := recovery.NewService()
	_, hashes1, err := svc.GenerateCodes(8)
	require.NoError(t, err)
	require.NoError(t, repo.CreateRecoveryCodes(ctx, user1.ID, hashes1))

	_, hashes2, err := svc.GenerateCodes(8)
	require.NoError(t, err)
	require.NoError(t, repo.CreateRecoveryCodes(ctx, user2.ID, hashes2))

	// Delete user1's codes
	err = repo.DeleteRecoveryCodes(ctx, user1.ID)
	require.NoError(t, err)

	// User1 should have 0, User2 should still have 8
	count1, err := repo.GetUnusedRecoveryCodeCount(ctx, user1.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count1)

	count2, err := repo.GetUnusedRecoveryCodeCount(ctx, user2.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(8), count2)
}

func TestHasRecoveryCodes(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser", "Test User")

	// Initially no codes
	has, err := repo.HasRecoveryCodes(ctx, user.ID)
	require.NoError(t, err)
	assert.False(t, has)

	// Add codes
	svc := recovery.NewService()
	_, hashes, err := svc.GenerateCodes(8)
	require.NoError(t, err)
	require.NoError(t, repo.CreateRecoveryCodes(ctx, user.ID, hashes))

	has, err = repo.HasRecoveryCodes(ctx, user.ID)
	require.NoError(t, err)
	assert.True(t, has)
}

func TestHasRecoveryCodes_IncludesUsedCodes(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)
	ctx := context.Background()

	user := testutil.NewTestUser(t, db, "testuser", "Test User")

	svc := recovery.NewService()
	plaintexts, hashes, err := svc.GenerateCodes(1)
	require.NoError(t, err)
	require.NoError(t, repo.CreateRecoveryCodes(ctx, user.ID, hashes))

	// Use the only code
	normalized := recovery.NormalizeCode(plaintexts[0])
	_, err = repo.ValidateAndUseRecoveryCode(ctx, user.ID, normalized)
	require.NoError(t, err)

	// Should still return true (has codes, even if used)
	has, err := repo.HasRecoveryCodes(ctx, user.ID)
	require.NoError(t, err)
	assert.True(t, has)
}
