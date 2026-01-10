// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package models

import "time"

// RecoveryCode stores a hashed recovery code for account recovery.
type RecoveryCode struct { //nolint:govet // fieldalignment not critical for this model
	ID        int64      `gorm:"primaryKey" json:"id"`
	UserID    int64      `gorm:"not null;index" json:"user_id"`
	CodeHash  string     `gorm:"not null" json:"-"`
	Used      bool       `gorm:"not null;default:false" json:"used"`
	CreatedAt time.Time  `json:"created_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}
