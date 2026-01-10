// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package recovery

import (
	"crypto/rand"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	// CodeLength is the length of each recovery code (without dashes).
	CodeLength = 12
	// CodeCount is the default number of recovery codes to generate.
	CodeCount = 8
	// bcryptCost is the cost factor for bcrypt hashing.
	bcryptCost = 10
)

// alphabet for recovery codes (lowercase + digits, excluding confusing chars: 0, o, l, 1).
const alphabet = "23456789abcdefghjkmnpqrstuvwxyz"

// Service handles recovery code generation and validation.
type Service struct{}

// NewService creates a new recovery service.
func NewService() *Service {
	return &Service{}
}

// GenerateCodes generates recovery codes and their hashes.
// Returns (plaintext codes for display, hashed codes for storage, error).
func (s *Service) GenerateCodes(count int) ([]string, []string, error) {
	if count <= 0 {
		count = CodeCount
	}

	plaintexts := make([]string, count)
	hashes := make([]string, count)

	for i := 0; i < count; i++ {
		code, err := generateCode(CodeLength)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate code: %w", err)
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(code), bcryptCost)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to hash code: %w", err)
		}

		plaintexts[i] = formatCode(code)
		hashes[i] = string(hash)
	}

	return plaintexts, hashes, nil
}

// NormalizeCode removes dashes and converts to lowercase for comparison.
func NormalizeCode(code string) string {
	code = strings.ReplaceAll(code, "-", "")
	return strings.ToLower(code)
}

// generateCode generates a random code of the specified length.
func generateCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i := range bytes {
		bytes[i] = alphabet[int(bytes[i])%len(alphabet)]
	}

	return string(bytes), nil
}

// formatCode formats a code with dashes for readability (e.g., "a1b2-c3d4-e5f6").
func formatCode(code string) string {
	var parts []string
	for i := 0; i < len(code); i += 4 {
		end := min(i+4, len(code))
		parts = append(parts, code[i:end])
	}
	return strings.Join(parts, "-")
}
