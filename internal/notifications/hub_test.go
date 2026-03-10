package notifications

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHub_AddAndRemoveClient(t *testing.T) {
	hub := NewHub()
	
	client1 := &Client{
		UserID: 1,
		Send:   make(chan interface{}, 1),
	}
	
	hub.AddClient(client1)
	
	hub.mu.RLock()
	assert.Len(t, hub.connections[1], 1)
	hub.mu.RUnlock()
	
	hub.RemoveClient(client1)
	
	hub.mu.RLock()
	assert.Nil(t, hub.connections[1])
	hub.mu.RUnlock()
	
	// Channel should be closed
	_, ok := <-client1.Send
	assert.False(t, ok)
}

func TestHub_SendToUser(t *testing.T) {
	hub := NewHub()
	
	client1 := &Client{
		UserID: 1,
		Send:   make(chan interface{}, 1),
	}
	
	hub.AddClient(client1)
	
	msg := map[string]string{"event": "test"}
	hub.SendToUser(1, msg)
	
	received := <-client1.Send
	assert.Equal(t, msg, received)
}
