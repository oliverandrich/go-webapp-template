// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package models

import (
	"time"
)

type User struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	IsAdmin      int64     `gorm:"not null;default:0" json:"is_admin"`
	CreatedAt    time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

// AllModels returns all models for GORM AutoMigrate
func AllModels() []any {
	return []any{
		&User{},
	}
}
