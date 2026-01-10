// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package recovery_test

import (
	"strings"
	"testing"

	"codeberg.org/oliverandrich/go-webapp-template/internal/services/recovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestNewService(t *testing.T) {
	svc := recovery.NewService()
	assert.NotNil(t, svc)
}

func TestGenerateCodes(t *testing.T) {
	svc := recovery.NewService()

	plaintexts, hashes, err := svc.GenerateCodes(8)

	require.NoError(t, err)
	assert.Len(t, plaintexts, 8)
	assert.Len(t, hashes, 8)
}

func TestGenerateCodes_DefaultCount(t *testing.T) {
	svc := recovery.NewService()

	plaintexts, hashes, err := svc.GenerateCodes(0)

	require.NoError(t, err)
	assert.Len(t, plaintexts, recovery.CodeCount)
	assert.Len(t, hashes, recovery.CodeCount)
}

func TestGenerateCodes_NegativeCount(t *testing.T) {
	svc := recovery.NewService()

	plaintexts, hashes, err := svc.GenerateCodes(-5)

	require.NoError(t, err)
	assert.Len(t, plaintexts, recovery.CodeCount)
	assert.Len(t, hashes, recovery.CodeCount)
}

func TestGenerateCodes_Format(t *testing.T) {
	svc := recovery.NewService()

	plaintexts, _, err := svc.GenerateCodes(1)

	require.NoError(t, err)
	code := plaintexts[0]

	// Code should be in format xxxx-xxxx-xxxx (12 chars + 2 dashes = 14 chars)
	assert.Len(t, code, 14)
	assert.Equal(t, '-', rune(code[4]))
	assert.Equal(t, '-', rune(code[9]))
}

func TestGenerateCodes_UniqueValues(t *testing.T) {
	svc := recovery.NewService()

	plaintexts, _, err := svc.GenerateCodes(100)

	require.NoError(t, err)

	// Check all codes are unique
	seen := make(map[string]bool)
	for _, code := range plaintexts {
		assert.False(t, seen[code], "Duplicate code found: %s", code)
		seen[code] = true
	}
}

func TestGenerateCodes_ValidCharacters(t *testing.T) {
	svc := recovery.NewService()

	plaintexts, _, err := svc.GenerateCodes(10)

	require.NoError(t, err)

	validChars := "23456789abcdefghjkmnpqrstuvwxyz-"
	for _, code := range plaintexts {
		for _, c := range code {
			assert.Contains(t, validChars, string(c), "Invalid character: %c", c)
		}
	}
}

func TestGenerateCodes_HashesMatchPlaintexts(t *testing.T) {
	svc := recovery.NewService()

	plaintexts, hashes, err := svc.GenerateCodes(5)

	require.NoError(t, err)

	for i, plaintext := range plaintexts {
		normalized := recovery.NormalizeCode(plaintext)
		err := bcrypt.CompareHashAndPassword([]byte(hashes[i]), []byte(normalized))
		assert.NoError(t, err, "Hash at index %d should match plaintext", i)
	}
}

func TestNormalizeCode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"a1b2-c3d4-e5f6", "a1b2c3d4e5f6"},
		{"A1B2-C3D4-E5F6", "a1b2c3d4e5f6"},
		{"a1b2c3d4e5f6", "a1b2c3d4e5f6"},
		{"A1B2C3D4E5F6", "a1b2c3d4e5f6"},
		{"", ""},
		{"----", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := recovery.NormalizeCode(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeCode_PreservesDigits(t *testing.T) {
	result := recovery.NormalizeCode("2345-6789-2345")
	assert.Equal(t, "234567892345", result)
}

func TestGenerateCodes_NoConfusingCharacters(t *testing.T) {
	svc := recovery.NewService()

	// Generate many codes to increase probability of catching bad chars
	plaintexts, _, err := svc.GenerateCodes(100)

	require.NoError(t, err)

	confusingChars := "0oOl1I"
	for _, code := range plaintexts {
		for _, c := range confusingChars {
			assert.False(t, strings.ContainsRune(code, c),
				"Code contains confusing character %c: %s", c, code)
		}
	}
}
