// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package sse

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHub_RegisterAndUnregister(t *testing.T) {
	hub := NewHub()

	// Register a client
	ch := hub.Register("session1", 1)
	assert.NotNil(t, ch)
	assert.Equal(t, 1, hub.ClientCount())
	assert.Equal(t, 1, hub.SessionCount())
	assert.Equal(t, 1, hub.UserCount())

	// Register another client for same session (e.g., second tab)
	ch2 := hub.Register("session1", 1)
	assert.Equal(t, 2, hub.ClientCount())
	assert.Equal(t, 1, hub.SessionCount()) // Still one session

	// Unregister first client
	hub.Unregister("session1", 1, ch)
	assert.Equal(t, 1, hub.ClientCount())
	assert.Equal(t, 1, hub.SessionCount())

	// Unregister second client
	hub.Unregister("session1", 1, ch2)
	assert.Equal(t, 0, hub.ClientCount())
	assert.Equal(t, 0, hub.SessionCount())
	assert.Equal(t, 0, hub.UserCount())
}

func TestHub_MultipleSessionsPerUser(t *testing.T) {
	hub := NewHub()

	// User 1 connects from two different browsers
	ch1 := hub.Register("session1", 1)
	ch2 := hub.Register("session2", 1)

	assert.Equal(t, 2, hub.ClientCount())
	assert.Equal(t, 2, hub.SessionCount())
	assert.Equal(t, 1, hub.UserCount()) // Still one user

	hub.Unregister("session1", 1, ch1)
	hub.Unregister("session2", 1, ch2)
}

//nolint:dupl // Test functions intentionally have similar structure for clarity
func TestHub_SendToSession(t *testing.T) {
	hub := NewHub()

	ch1 := hub.Register("session1", 1)
	ch2 := hub.Register("session1", 1) // Same session, different tab
	ch3 := hub.Register("session2", 2) // Different session

	hub.SendToSession("session1", "hello")

	// Both clients in session1 should receive the message
	select {
	case msg := <-ch1:
		assert.Equal(t, "hello", msg)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("ch1 should have received message")
	}

	select {
	case msg := <-ch2:
		assert.Equal(t, "hello", msg)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("ch2 should have received message")
	}

	// Client in session2 should not receive the message
	select {
	case <-ch3:
		t.Fatal("ch3 should not have received message")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}

	hub.Unregister("session1", 1, ch1)
	hub.Unregister("session1", 1, ch2)
	hub.Unregister("session2", 2, ch3)
}

//nolint:dupl // Test functions intentionally have similar structure for clarity
func TestHub_SendToUser(t *testing.T) {
	hub := NewHub()

	// User 1 has two sessions
	ch1 := hub.Register("session1", 1)
	ch2 := hub.Register("session2", 1)
	// User 2 has one session
	ch3 := hub.Register("session3", 2)

	hub.SendToUser(1, "user1-message")

	// Both sessions of user 1 should receive the message
	select {
	case msg := <-ch1:
		assert.Equal(t, "user1-message", msg)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("ch1 should have received message")
	}

	select {
	case msg := <-ch2:
		assert.Equal(t, "user1-message", msg)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("ch2 should have received message")
	}

	// User 2 should not receive the message
	select {
	case <-ch3:
		t.Fatal("ch3 should not have received message")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}

	hub.Unregister("session1", 1, ch1)
	hub.Unregister("session2", 1, ch2)
	hub.Unregister("session3", 2, ch3)
}

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub()

	ch1 := hub.Register("session1", 1)
	ch2 := hub.Register("session2", 2)

	hub.Broadcast("broadcast-message")

	// All clients should receive the message
	select {
	case msg := <-ch1:
		assert.Equal(t, "broadcast-message", msg)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("ch1 should have received message")
	}

	select {
	case msg := <-ch2:
		assert.Equal(t, "broadcast-message", msg)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("ch2 should have received message")
	}

	hub.Unregister("session1", 1, ch1)
	hub.Unregister("session2", 2, ch2)
}

func TestHub_NonBlockingSend(t *testing.T) {
	hub := NewHub()

	ch := hub.Register("session1", 1)

	// Fill the channel buffer (size 10)
	for range 10 {
		hub.SendToSession("session1", "msg")
	}

	// This should not block even though buffer is full
	done := make(chan bool)
	go func() {
		hub.SendToSession("session1", "overflow")
		done <- true
	}()

	select {
	case <-done:
		// Expected - send should not block
	case <-time.After(100 * time.Millisecond):
		t.Fatal("SendToSession blocked on full channel")
	}

	hub.Unregister("session1", 1, ch)
}

func TestHub_ConcurrentAccess(t *testing.T) {
	hub := NewHub()

	var wg sync.WaitGroup
	const numGoroutines = 100

	// Concurrent registrations
	channels := make([]chan string, numGoroutines)
	for i := range numGoroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			channels[idx] = hub.Register("session", 1)
		}(i)
	}
	wg.Wait()

	assert.Equal(t, numGoroutines, hub.ClientCount())

	// Concurrent sends
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hub.SendToSession("session", "concurrent")
		}()
	}
	wg.Wait()

	// Concurrent unregistrations
	for i := range numGoroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			hub.Unregister("session", 1, channels[idx])
		}(i)
	}
	wg.Wait()

	assert.Equal(t, 0, hub.ClientCount())
}
