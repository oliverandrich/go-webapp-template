// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

// Package htmx provides types and helpers for htmx integration.
package htmx

import (
	"net/http"
)

// Header constants for htmx request headers.
const (
	HeaderRequest        = "HX-Request"
	HeaderBoosted        = "HX-Boosted"
	HeaderCurrentURL     = "HX-Current-URL"
	HeaderHistoryRestore = "HX-History-Restore-Request"
	HeaderPrompt         = "HX-Prompt"
	HeaderTarget         = "HX-Target"
	HeaderTrigger        = "HX-Trigger"
	HeaderTriggerName    = "HX-Trigger-Name"
)

// Header constants for htmx response headers.
const (
	HeaderLocation           = "HX-Location"
	HeaderPushURL            = "HX-Push-Url"
	HeaderRedirect           = "HX-Redirect"
	HeaderRefresh            = "HX-Refresh"
	HeaderReplaceURL         = "HX-Replace-Url"
	HeaderReswap             = "HX-Reswap"
	HeaderRetarget           = "HX-Retarget"
	HeaderReselect           = "HX-Reselect"
	HeaderTriggerResponse    = "HX-Trigger"
	HeaderTriggerAfterSettle = "HX-Trigger-After-Settle"
	HeaderTriggerAfterSwap   = "HX-Trigger-After-Swap"
)

// Request contains information about an htmx request.
// Similar to django-htmx's HtmxDetails.
type Request struct { //nolint:govet // fieldalignment not critical
	// IsHtmx is true if this is an htmx request (HX-Request header is "true").
	IsHtmx bool

	// IsBoosted is true if this is a boosted request (HX-Boosted header is "true").
	IsBoosted bool

	// CurrentURL is the current URL of the browser (HX-Current-URL header).
	CurrentURL string

	// IsHistoryRestore is true if this is a history restore request.
	IsHistoryRestore bool

	// Prompt is the user response to an hx-prompt (HX-Prompt header).
	Prompt string

	// Target is the ID of the target element (HX-Target header).
	Target string

	// Trigger is the ID of the triggered element (HX-Trigger header).
	Trigger string

	// TriggerName is the name of the triggered element (HX-Trigger-Name header).
	TriggerName string
}

// ParseRequest extracts htmx information from request headers.
func ParseRequest(r *http.Request) *Request {
	return &Request{
		IsHtmx:           r.Header.Get(HeaderRequest) == "true",
		IsBoosted:        r.Header.Get(HeaderBoosted) == "true",
		CurrentURL:       r.Header.Get(HeaderCurrentURL),
		IsHistoryRestore: r.Header.Get(HeaderHistoryRestore) == "true",
		Prompt:           r.Header.Get(HeaderPrompt),
		Target:           r.Header.Get(HeaderTarget),
		Trigger:          r.Header.Get(HeaderTrigger),
		TriggerName:      r.Header.Get(HeaderTriggerName),
	}
}
