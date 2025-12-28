// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package middleware

import (
	"net/http"

	"codeberg.org/oliverandrich/go-webapp-template/internal/i18n"
)

// Locale creates middleware that detects the user's preferred language
// from the Accept-Language header and sets it in the request context.
func Locale(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptLang := r.Header.Get("Accept-Language")
		lang := i18n.MatchLanguage(acceptLang)
		ctx := i18n.WithLocale(r.Context(), lang)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
