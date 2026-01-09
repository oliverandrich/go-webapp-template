// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package htmx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"codeberg.org/oliverandrich/go-webapp-template/internal/htmx"
	"github.com/stretchr/testify/assert"
)

func TestParseRequest_HtmxRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("HX-Request", "true")

	parsed := htmx.ParseRequest(req)

	assert.True(t, parsed.IsHtmx)
}

func TestParseRequest_NonHtmxRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	parsed := htmx.ParseRequest(req)

	assert.False(t, parsed.IsHtmx)
}

func TestParseRequest_AllHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Boosted", "true")
	req.Header.Set("HX-Current-URL", "http://example.com/page")
	req.Header.Set("HX-History-Restore-Request", "true")
	req.Header.Set("HX-Prompt", "user input")
	req.Header.Set("HX-Target", "target-id")
	req.Header.Set("HX-Trigger", "trigger-id")
	req.Header.Set("HX-Trigger-Name", "trigger-name")

	parsed := htmx.ParseRequest(req)

	assert.True(t, parsed.IsHtmx)
	assert.True(t, parsed.IsBoosted)
	assert.Equal(t, "http://example.com/page", parsed.CurrentURL)
	assert.True(t, parsed.IsHistoryRestore)
	assert.Equal(t, "user input", parsed.Prompt)
	assert.Equal(t, "target-id", parsed.Target)
	assert.Equal(t, "trigger-id", parsed.Trigger)
	assert.Equal(t, "trigger-name", parsed.TriggerName)
}

func TestParseRequest_BoostedNotHtmx(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("HX-Boosted", "true")
	// Note: HX-Request not set, so IsHtmx should be false

	parsed := htmx.ParseRequest(req)

	assert.False(t, parsed.IsHtmx)
	assert.True(t, parsed.IsBoosted)
}

func TestParseRequest_FalseValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("HX-Request", "false")
	req.Header.Set("HX-Boosted", "false")

	parsed := htmx.ParseRequest(req)

	assert.False(t, parsed.IsHtmx)
	assert.False(t, parsed.IsBoosted)
}
