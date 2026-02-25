package notifications

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	connections map[int][]*websocket.Conn
	mu          sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		connections: make(map[int][]*websocket.Conn),
	}
}

func (h *Hub) AddConnection(userID int, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[userID] = append(h.connections[userID], conn)
}

func (h *Hub) RemoveConnection(userID int, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conns := h.connections[userID]
	for i, c := range conns {
		if c == conn {
			h.connections[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
	if len(h.connections[userID]) == 0 {
		delete(h.connections, userID)
	}
}

func (h *Hub) SendToUser(userID int, message interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if conns, ok := h.connections[userID]; ok {
		for _, conn := range conns {
			conn.WriteJSON(message)
		}
	}
}
