// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

// Package htmx provides helper functions for HTMX-aware request handling.
package htmx

import "net/http"

// Header constants for HTMX requests and responses.
const (
	HeaderRequest    = "HX-Request"
	HeaderTrigger    = "HX-Trigger"
	HeaderRedirect   = "HX-Redirect"
	HeaderRetarget   = "HX-Retarget"
	HeaderReswap     = "HX-Reswap"
	HeaderPushURL    = "HX-Push-Url"
	HeaderRefresh    = "HX-Refresh"
	HeaderReplaceURL = "HX-Replace-Url"
)

// IsRequest returns true if the request is an HTMX request.
func IsRequest(r *http.Request) bool {
	return r.Header.Get(HeaderRequest) == "true"
}

// Redirect performs an HTMX-aware redirect.
func Redirect(w http.ResponseWriter, r *http.Request, url string) {
	if IsRequest(r) {
		w.Header().Set(HeaderRedirect, url)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// TriggerEvent sends an HX-Trigger header to trigger client-side events.
func TriggerEvent(w http.ResponseWriter, event string) {
	w.Header().Set(HeaderTrigger, event)
}

// Retarget changes the target element for the response.
func Retarget(w http.ResponseWriter, selector string) {
	w.Header().Set(HeaderRetarget, selector)
}

// Reswap changes the swap method for the response.
func Reswap(w http.ResponseWriter, swapMethod string) {
	w.Header().Set(HeaderReswap, swapMethod)
}

// PushURL instructs HTMX to push a new URL into the browser history.
func PushURL(w http.ResponseWriter, url string) {
	w.Header().Set(HeaderPushURL, url)
}

// ReplaceURL instructs HTMX to replace the current URL in browser history.
func ReplaceURL(w http.ResponseWriter, url string) {
	w.Header().Set(HeaderReplaceURL, url)
}

// Refresh instructs HTMX to do a full page refresh.
func Refresh(w http.ResponseWriter) {
	w.Header().Set(HeaderRefresh, "true")
}
