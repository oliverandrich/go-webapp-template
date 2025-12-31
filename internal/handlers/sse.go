// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package handlers

import (
	"net/http"
	"time"

	"github.com/oliverandrich/go-webapp-template/internal/services/session"
	"github.com/oliverandrich/go-webapp-template/internal/sse"
	"github.com/alexedwards/scs/v2"
)

// SSEHandler handles Server-Sent Events connections.
type SSEHandler struct {
	hub            *sse.Hub
	sessionManager *scs.SessionManager
}

// NewSSEHandler creates a new SSE handler.
func NewSSEHandler(hub *sse.Hub, sessionManager *scs.SessionManager) *SSEHandler {
	return &SSEHandler{
		hub:            hub,
		sessionManager: sessionManager,
	}
}

// Events handles the SSE connection endpoint.
func (h *SSEHandler) Events(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get session token (used as session ID)
	sessionToken := h.sessionManager.Token(ctx)
	if sessionToken == "" {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	// Get user from session
	user := session.GetUser(ctx, h.sessionManager)
	if user == nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	// Check if response supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Register client with session and user ID
	ch := h.hub.Register(sessionToken, user.ID)
	defer h.hub.Unregister(sessionToken, user.ID, ch)

	// Send initial connection event
	w.Write([]byte(sse.FormatEvent("connected", "ok")))
	flusher.Flush()

	// Heartbeat ticker to keep connection alive through proxies
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Stream events until client disconnects
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := w.Write([]byte(sse.Heartbeat)); err != nil {
				return // Client disconnected
			}
			flusher.Flush()
		case msg, ok := <-ch:
			if !ok {
				return
			}
			w.Write([]byte(msg))
			flusher.Flush()
		}
	}
}

// Hub returns the SSE hub for sending messages from other handlers.
func (h *SSEHandler) Hub() *sse.Hub {
	return h.hub
}
