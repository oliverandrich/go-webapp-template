// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package i18n_test

import (
	"context"
	"testing"

	"github.com/oliverandrich/go-webapp-template/internal/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func TestInit(t *testing.T) {
	err := i18n.Init()
	require.NoError(t, err)
}

func TestT(t *testing.T) {
	require.NoError(t, i18n.Init())

	ctx := i18n.WithLocale(context.Background(), language.English)

	// Should translate known key
	result := i18n.T(ctx, "app_name")
	assert.NotEqual(t, "app_name", result) // Should be translated
}

func TestT_German(t *testing.T) {
	require.NoError(t, i18n.Init())

	ctx := i18n.WithLocale(context.Background(), language.German)

	// Should use German translation
	result := i18n.T(ctx, "app_name")
	assert.NotEqual(t, "app_name", result)
}

func TestT_UnknownKey(t *testing.T) {
	require.NoError(t, i18n.Init())

	ctx := i18n.WithLocale(context.Background(), language.English)

	// Should return the key itself for unknown messages
	result := i18n.T(ctx, "unknown_key_that_does_not_exist")
	assert.Equal(t, "unknown_key_that_does_not_exist", result)
}

func TestT_NoLocaleContext(t *testing.T) {
	require.NoError(t, i18n.Init())

	// Without WithLocale, should fallback to English
	ctx := context.Background()

	result := i18n.T(ctx, "app_name")
	assert.NotEmpty(t, result)
}

func TestTData(t *testing.T) {
	require.NoError(t, i18n.Init())

	ctx := i18n.WithLocale(context.Background(), language.English)

	// Test with template data (if we have any messages that use it)
	// For now, just verify it doesn't panic
	result := i18n.TData(ctx, "app_name", map[string]any{"Name": "Test"})
	assert.NotEmpty(t, result)
}

func TestTPlural(t *testing.T) {
	require.NoError(t, i18n.Init())

	ctx := i18n.WithLocale(context.Background(), language.English)

	// Test plural - will return key if no plural message defined
	result := i18n.TPlural(ctx, "app_name", 1)
	assert.NotEmpty(t, result)

	result = i18n.TPlural(ctx, "app_name", 5)
	assert.NotEmpty(t, result)
}

func TestMatchLanguage(t *testing.T) {
	tests := []struct {
		expected       language.Tag
		acceptLanguage string
	}{
		{language.English, "en"},
		{language.English, "en-US"},
		{language.German, "de"},
		{language.German, "de-DE"},
		{language.German, "de-AT"},
		{language.English, "fr"}, // fallback to English
		{language.English, ""},   // empty defaults to English
		{language.German, "de, en;q=0.9"},
		{language.English, "en, de;q=0.9"},
	}

	for _, tt := range tests {
		t.Run(tt.acceptLanguage, func(t *testing.T) {
			tag := i18n.MatchLanguage(tt.acceptLanguage)
			// Compare base language (ignore region)
			assert.Equal(t, tt.expected.String()[:2], tag.String()[:2])
		})
	}
}

func TestWithLocale(t *testing.T) {
	require.NoError(t, i18n.Init())

	ctx := i18n.WithLocale(context.Background(), language.German)

	locale := i18n.GetLocale(ctx)
	assert.Equal(t, "de", locale)
}

func TestGetLocale(t *testing.T) {
	require.NoError(t, i18n.Init())

	ctx := i18n.WithLocale(context.Background(), language.English)

	assert.Equal(t, "en", i18n.GetLocale(ctx))
}

func TestGetLocale_Default(t *testing.T) {
	ctx := context.Background()

	// Without WithLocale, should return "en"
	assert.Equal(t, "en", i18n.GetLocale(ctx))
}
