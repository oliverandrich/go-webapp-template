// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository_test

import (
	"testing"

	"codeberg.org/oliverandrich/go-webapp-template/internal/repository"
	"codeberg.org/oliverandrich/go-webapp-template/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := repository.New(db)

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.DB())
}
