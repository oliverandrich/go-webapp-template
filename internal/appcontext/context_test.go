// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package appcontext_test

import (
	"testing"

	"codeberg.org/oliverandrich/go-webapp-template/internal/appcontext"
	"codeberg.org/oliverandrich/go-webapp-template/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestContext_GetUser(t *testing.T) {
	user := &models.User{ID: 123, Username: "testuser"}
	ctx := &appcontext.Context{User: user}

	result := ctx.GetUser()

	assert.Equal(t, user, result)
	assert.Equal(t, int64(123), result.ID)
}

func TestContext_GetUser_Nil(t *testing.T) {
	ctx := &appcontext.Context{User: nil}

	result := ctx.GetUser()

	assert.Nil(t, result)
}

func TestContext_IsAuthenticated_True(t *testing.T) {
	ctx := &appcontext.Context{User: &models.User{ID: 1}}

	assert.True(t, ctx.IsAuthenticated())
}

func TestContext_IsAuthenticated_False(t *testing.T) {
	ctx := &appcontext.Context{User: nil}

	assert.False(t, ctx.IsAuthenticated())
}
