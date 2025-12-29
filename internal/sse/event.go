// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package sse

import (
	"fmt"
	"strings"
)

// FormatEvent formats a message as an SSE event with optional event name.
// Multiline content is properly prefixed with "data:".
func FormatEvent(eventName, data string) string {
	var sb strings.Builder

	if eventName != "" {
		sb.WriteString(fmt.Sprintf("event: %s\n", eventName))
	}

	// Handle multiline data
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		sb.WriteString(fmt.Sprintf("data: %s\n", line))
	}

	sb.WriteString("\n") // Empty line marks end of event
	return sb.String()
}

// FormatOOBEvent formats an HTMX OOB swap event.
// The HTML should include hx-swap-oob attribute.
func FormatOOBEvent(html string) string {
	return FormatEvent("message", html)
}

// FormatNamedOOBEvent formats an HTMX OOB swap event with a custom event name.
func FormatNamedOOBEvent(eventName, html string) string {
	return FormatEvent(eventName, html)
}

// Heartbeat is an SSE comment that keeps the connection alive.
// Comments (lines starting with :) are ignored by SSE clients.
const Heartbeat = ": heartbeat\n\n"
