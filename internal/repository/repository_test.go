// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository_test

import (
	"testing"

	"github.com/oliverandrich/go-webapp-template/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	_, repo := testutil.NewTestDB(t)

	assert.NotNil(t, repo)
}
