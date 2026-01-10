// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package models

import "time"

// EmailVerificationToken stores a hashed token for email verification.
type EmailVerificationToken struct { //nolint:govet // fieldalignment not critical
	ID        int64     `gorm:"primaryKey" json:"id"`
	UserID    int64     `gorm:"not null;index" json:"user_id"`
	TokenHash string    `gorm:"uniqueIndex;not null" json:"-"` // SHA256 hash
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
