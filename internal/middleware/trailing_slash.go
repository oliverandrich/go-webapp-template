// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package middleware

import (
	"net/http"
	"strings"
)

// StripTrailingSlash redirects requests with trailing slashes to the canonical URL without.
func StripTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path != "/" && strings.HasSuffix(path, "/") {
			newPath := strings.TrimSuffix(path, "/")
			newURL := newPath
			if r.URL.RawQuery != "" {
				newURL += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, newURL, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}
