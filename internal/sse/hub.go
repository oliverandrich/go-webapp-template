// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package sse

import (
	"sync"

	"github.com/samber/lo"
)

// client represents a connected SSE client with its channel and user association.
type client struct {
	ch     chan string
	userID int64
}

// Hub manages SSE clients per session and user.
// Multiple tabs/windows from the same session share the same session ID.
// Multiple sessions can belong to the same user (different browsers).
type Hub struct {
	clients      map[string][]client
	userSessions map[int64][]string
	mu           sync.RWMutex
}

// NewHub creates a new SSE hub.
func NewHub() *Hub {
	return &Hub{
		clients:      make(map[string][]client),
		userSessions: make(map[int64][]string),
	}
}

// Register adds a new client channel for the given session and user.
// Returns the channel to receive events on.
func (h *Hub) Register(sessionID string, userID int64) chan string {
	ch := make(chan string, 10) // buffered to prevent blocking

	h.mu.Lock()
	defer h.mu.Unlock()

	// Add client to session
	h.clients[sessionID] = append(h.clients[sessionID], client{ch: ch, userID: userID})

	// Track session for user (if not already tracked)
	if !lo.Contains(h.userSessions[userID], sessionID) {
		h.userSessions[userID] = append(h.userSessions[userID], sessionID)
	}

	return ch
}

// Unregister removes a client channel for the given session.
func (h *Hub) Unregister(sessionID string, userID int64, ch chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove client from session
	clients := h.clients[sessionID]
	h.clients[sessionID] = lo.Filter(clients, func(c client, _ int) bool {
		return c.ch != ch
	})

	// Clean up empty session entries
	if len(h.clients[sessionID]) == 0 {
		delete(h.clients, sessionID)

		// Remove session from user's session list
		h.userSessions[userID] = lo.Filter(h.userSessions[userID], func(s string, _ int) bool {
			return s != sessionID
		})
		if len(h.userSessions[userID]) == 0 {
			delete(h.userSessions, userID)
		}
	}

	close(ch)
}

// SendToSession sends a message to all clients of the given session.
func (h *Hub) SendToSession(sessionID string, message string) {
	h.mu.RLock()
	clients := h.clients[sessionID]
	h.mu.RUnlock()

	for _, c := range clients {
		select {
		case c.ch <- message:
		default:
			// Channel full, skip (prevents blocking)
		}
	}
}

// SendToUser sends a message to all sessions of the given user.
// This reaches the user across all browsers/devices.
func (h *Hub) SendToUser(userID int64, message string) {
	h.mu.RLock()
	sessionIDs := h.userSessions[userID]
	h.mu.RUnlock()

	for _, sessionID := range sessionIDs {
		h.SendToSession(sessionID, message)
	}
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(message string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, clients := range h.clients {
		for _, c := range clients {
			select {
			case c.ch <- message:
			default:
				// Channel full, skip
			}
		}
	}
}

// ClientCount returns the total number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return lo.SumBy(lo.Values(h.clients), func(clients []client) int {
		return len(clients)
	})
}

// SessionCount returns the number of unique sessions with active connections.
func (h *Hub) SessionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.clients)
}

// UserCount returns the number of unique users with active connections.
func (h *Hub) UserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.userSessions)
}
