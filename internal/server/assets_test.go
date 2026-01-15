// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package server

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindAssets(t *testing.T) {
	assets := findAssets()

	// CSSPath should be in /static/dist/ and end with .css
	assert.True(t, strings.HasPrefix(assets.CSSPath, "/static/dist/styles"), "CSSPath should start with /static/dist/styles")
	assert.True(t, strings.HasSuffix(assets.CSSPath, ".css"), "CSSPath should end with .css")

	// JSPath should be in /static/dist/ and end with .js
	assert.True(t, strings.HasPrefix(assets.JSPath, "/static/dist/app"), "JSPath should start with /static/dist/app")
	assert.True(t, strings.HasSuffix(assets.JSPath, ".js"), "JSPath should end with .js")
}
