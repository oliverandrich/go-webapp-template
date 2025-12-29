// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package sse

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventName string
		data      string
		expected  string
	}{
		{
			name:      "simple message without event name",
			eventName: "",
			data:      "hello",
			expected:  "data: hello\n\n",
		},
		{
			name:      "simple message with event name",
			eventName: "update",
			data:      "hello",
			expected:  "event: update\ndata: hello\n\n",
		},
		{
			name:      "multiline data",
			eventName: "",
			data:      "line1\nline2\nline3",
			expected:  "data: line1\ndata: line2\ndata: line3\n\n",
		},
		{
			name:      "multiline data with event name",
			eventName: "update",
			data:      "line1\nline2",
			expected:  "event: update\ndata: line1\ndata: line2\n\n",
		},
		{
			name:      "HTML content",
			eventName: "message",
			data:      `<div id="test">Hello</div>`,
			expected:  "event: message\ndata: <div id=\"test\">Hello</div>\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatEvent(tt.eventName, tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatOOBEvent(t *testing.T) {
	html := `<div id="counter" hx-swap-oob="true">42</div>`
	result := FormatOOBEvent(html)

	expected := "event: message\ndata: <div id=\"counter\" hx-swap-oob=\"true\">42</div>\n\n"
	assert.Equal(t, expected, result)
}

func TestFormatNamedOOBEvent(t *testing.T) {
	html := `<div id="notification">New message!</div>`
	result := FormatNamedOOBEvent("notify", html)

	expected := "event: notify\ndata: <div id=\"notification\">New message!</div>\n\n"
	assert.Equal(t, expected, result)
}

func TestHeartbeat(t *testing.T) {
	// Heartbeat should be a valid SSE comment
	assert.Equal(t, ": heartbeat\n\n", Heartbeat)
	// Should start with colon (SSE comment)
	assert.Equal(t, ':', rune(Heartbeat[0]))
}
