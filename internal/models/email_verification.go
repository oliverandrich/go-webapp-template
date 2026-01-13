// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package models

import "time"

// EmailVerificationToken stores a hashed token for email verification.
type EmailVerificationToken struct { //nolint:govet // fieldalignment: readability over optimization
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	TokenHash string    `db:"token_hash" json:"-"` // SHA256 hash
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
