// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package repository

import (
	"gorm.io/gorm"
)

// Repository provides data access methods.
type Repository struct {
	db *gorm.DB
}

// New creates a new Repository.
func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// DB returns the underlying database connection.
func (r *Repository) DB() *gorm.DB {
	return r.db
}
