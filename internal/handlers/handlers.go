// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package handlers

import (
	"encoding/json"
	"net/http"

	"codeberg.org/oliverandrich/go-webapp-template/internal/repository"
	"codeberg.org/oliverandrich/go-webapp-template/templates/home"
)

type Handler struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	home.Index().Render(r.Context(), w)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
