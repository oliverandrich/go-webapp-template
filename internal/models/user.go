// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package models

import (
	"time"
)

type User struct {
	CreatedAt    time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	IsAdmin      int64     `gorm:"not null;default:0" json:"is_admin"`
}

// AllModels returns all models for GORM AutoMigrate
func AllModels() []any {
	return []any{
		&User{},
	}
}
