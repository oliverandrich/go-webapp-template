// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package handlers

import (
	"fmt"
	"net/http"

	"codeberg.org/oliverandrich/go-webapp-template/templates/layouts"
)

// NotFound renders the 404 error page.
func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	layouts.Error404().Render(r.Context(), w)
}

// Forbidden renders the 403 error page.
func Forbidden(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	layouts.Error403().Render(r.Context(), w)
}

// RenderError renders a generic error page with the given status code and message.
func RenderError(w http.ResponseWriter, r *http.Request, code int, message string) {
	w.WriteHeader(code)

	codeStr := fmt.Sprintf("%d", code)
	title := http.StatusText(code)
	if title == "" {
		title = "Error"
	}

	layouts.Error(codeStr, title, message).Render(r.Context(), w)
}

// BadRequest renders a 400 error page.
func BadRequest(w http.ResponseWriter, r *http.Request, message string) {
	RenderError(w, r, http.StatusBadRequest, message)
}

// InternalServerError renders a 500 error page.
func InternalServerError(w http.ResponseWriter, r *http.Request, message string) {
	RenderError(w, r, http.StatusInternalServerError, message)
}

// Unauthorized renders a 401 error page.
func Unauthorized(w http.ResponseWriter, r *http.Request, message string) {
	RenderError(w, r, http.StatusUnauthorized, message)
}
