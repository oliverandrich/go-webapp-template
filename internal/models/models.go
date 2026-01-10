// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package models

// AllModels returns all models for database migration.
func AllModels() []any {
	return []any{
		&User{},
		&Credential{},
		&RecoveryCode{},
		&EmailVerificationToken{},
	}
}
