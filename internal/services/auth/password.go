// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package auth

import (
	"bufio"
	"embed"
	"fmt"
	"strings"
	"unicode"
)

//go:embed common_passwords.txt
var commonPasswordsFS embed.FS

var commonPasswords map[string]struct{}

func init() {
	commonPasswords = make(map[string]struct{})
	file, err := commonPasswordsFS.Open("common_passwords.txt")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		password := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if password != "" {
			commonPasswords[password] = struct{}{}
		}
	}
}

// PasswordValidator validates passwords against various criteria
type PasswordValidator struct {
	MinLength            int
	RequireUppercase     bool
	RequireLowercase     bool
	RequireDigit         bool
	RequireSpecial       bool
	CheckCommonPasswords bool
	CheckUserSimilarity  bool
}

// DefaultPasswordValidator returns a validator with sensible defaults
func DefaultPasswordValidator() *PasswordValidator {
	return &PasswordValidator{
		MinLength:            12,
		RequireUppercase:     false,
		RequireLowercase:     false,
		RequireDigit:         false,
		RequireSpecial:       false,
		CheckCommonPasswords: true,
		CheckUserSimilarity:  true,
	}
}

// ValidationError represents a single password validation error
type ValidationError struct {
	Code    string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// PasswordValidationError wraps multiple validation errors
type PasswordValidationError struct {
	Errors []ValidationError
}

func (e *PasswordValidationError) Error() string {
	if len(e.Errors) == 0 {
		return "password validation failed"
	}
	return e.Errors[0].Message
}

// Messages returns all error messages
func (e *PasswordValidationError) Messages() []string {
	messages := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		messages[i] = err.Message
	}
	return messages
}

// ValidationResult holds all validation errors
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// Validate checks a password against all configured validators
func (v *PasswordValidator) Validate(password string, userAttributes ...string) ValidationResult {
	var errors []ValidationError

	// Minimum length check
	if len(password) < v.MinLength {
		errors = append(errors, ValidationError{
			Code:    "min_length",
			Message: fmt.Sprintf("Password must be at least %d characters long.", v.MinLength),
		})
	}

	// Character type checks
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if v.RequireUppercase && !hasUpper {
		errors = append(errors, ValidationError{
			Code:    "no_uppercase",
			Message: "Password must contain at least one uppercase letter.",
		})
	}

	if v.RequireLowercase && !hasLower {
		errors = append(errors, ValidationError{
			Code:    "no_lowercase",
			Message: "Password must contain at least one lowercase letter.",
		})
	}

	if v.RequireDigit && !hasDigit {
		errors = append(errors, ValidationError{
			Code:    "no_digit",
			Message: "Password must contain at least one digit.",
		})
	}

	if v.RequireSpecial && !hasSpecial {
		errors = append(errors, ValidationError{
			Code:    "no_special",
			Message: "Password must contain at least one special character.",
		})
	}

	// Entirely numeric check
	if isEntirelyNumeric(password) {
		errors = append(errors, ValidationError{
			Code:    "entirely_numeric",
			Message: "Password cannot be entirely numeric.",
		})
	}

	// Common password check
	if v.CheckCommonPasswords && isCommonPassword(password) {
		errors = append(errors, ValidationError{
			Code:    "common_password",
			Message: "This password is too common. Please choose a more secure password.",
		})
	}

	// User attribute similarity check
	if v.CheckUserSimilarity && len(userAttributes) > 0 {
		if isSimilarToUserAttributes(password, userAttributes) {
			errors = append(errors, ValidationError{
				Code:    "too_similar",
				Message: "Password is too similar to your personal information.",
			})
		}
	}

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// GetHelpTexts returns help texts for password requirements
func (v *PasswordValidator) GetHelpTexts() []string {
	var texts []string

	texts = append(texts, fmt.Sprintf("At least %d characters", v.MinLength))

	if v.RequireUppercase {
		texts = append(texts, "At least one uppercase letter")
	}
	if v.RequireLowercase {
		texts = append(texts, "At least one lowercase letter")
	}
	if v.RequireDigit {
		texts = append(texts, "At least one digit")
	}
	if v.RequireSpecial {
		texts = append(texts, "At least one special character")
	}

	texts = append(texts, "Cannot be entirely numeric")

	if v.CheckCommonPasswords {
		texts = append(texts, "Not a commonly used password")
	}
	if v.CheckUserSimilarity {
		texts = append(texts, "Not too similar to your personal information")
	}

	return texts
}

func isEntirelyNumeric(password string) bool {
	for _, r := range password {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(password) > 0
}

func isCommonPassword(password string) bool {
	_, exists := commonPasswords[strings.ToLower(password)]
	return exists
}

func isSimilarToUserAttributes(password string, attributes []string) bool {
	passwordLower := strings.ToLower(password)

	for _, attr := range attributes {
		if attr == "" {
			continue
		}
		attrLower := strings.ToLower(attr)

		// Check if password contains the attribute
		if strings.Contains(passwordLower, attrLower) {
			return true
		}

		// Check if attribute contains the password
		if strings.Contains(attrLower, passwordLower) {
			return true
		}

		// Check similarity using a simple ratio
		if similarity(passwordLower, attrLower) > 0.7 {
			return true
		}
	}

	return false
}

func similarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	lcs := longestCommonSubsequence(a, b)
	maxLen := max(len(a), len(b))

	return float64(lcs) / float64(maxLen)
}

func longestCommonSubsequence(a, b string) int {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	return dp[m][n]
}
