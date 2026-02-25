package notifications

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Hub    *Hub
	UserID int
	Conn   *websocket.Conn
	Send   chan interface{}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()
	for {
		message, ok := <-c.Send
		if !ok {
			// The hub closed the channel.
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		c.Conn.WriteJSON(message)
	}
}

type Hub struct {
	connections map[int]map[*Client]bool
	mu          sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		connections: make(map[int]map[*Client]bool),
	}
}

func (h *Hub) AddClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.connections[client.UserID] == nil {
		h.connections[client.UserID] = make(map[*Client]bool)
	}
	h.connections[client.UserID][client] = true
}

func (h *Hub) RemoveClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.connections[client.UserID]; ok {
		if _, exists := conns[client]; exists {
			delete(conns, client)
			close(client.Send) // Signal WritePump to exit
			if len(conns) == 0 {
				delete(h.connections, client.UserID)
			}
		}
	}
}

func (h *Hub) SendToUser(userID int, message interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if conns, ok := h.connections[userID]; ok {
		for client := range conns {
			// Send message asynchronously over channel without blocking the lock
			select {
			case client.Send <- message:
			default:
				// If the client's channel is blocked/full, we skip or handle it
			}
		}
	}
}

