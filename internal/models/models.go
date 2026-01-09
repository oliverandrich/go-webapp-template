// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package models

import (
	"time"
)

// Example is a sample model to demonstrate GORM usage.
// Replace or extend with your own models.
type Example struct { //nolint:govet // fieldalignment not critical for models
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AllModels returns all models for database migration.
func AllModels() []any {
	return []any{
		&Example{},
	}
}
