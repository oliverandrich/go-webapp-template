// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

// Package repository provides data access using sqlx with direct model mapping.
package repository

import (
	"github.com/vinovest/sqlx"
)

// Repository provides data access methods.
type Repository struct {
	db *sqlx.DB
}

// New creates a new Repository.
func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}
