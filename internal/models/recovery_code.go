// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package models

import "time"

// RecoveryCode stores a hashed recovery code for account recovery.
type RecoveryCode struct { //nolint:govet // fieldalignment: readability over optimization
	ID        int64      `db:"id" json:"id"`
	UserID    int64      `db:"user_id" json:"user_id"`
	CodeHash  string     `db:"code_hash" json:"-"`
	Used      bool       `db:"used" json:"used"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UsedAt    *time.Time `db:"used_at" json:"used_at,omitempty"`
}
