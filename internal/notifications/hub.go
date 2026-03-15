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

type Hub struct {
	connections    map[int]map[*Client]bool
	mutex          sync.RWMutex
}

// --- Public ---

func NewHub() *Hub {
	return &Hub{
		connections: make(map[int]map[*Client]bool),
	}
}

func (clientInstance *Client) WritePump() {
	defer clientInstance.Conn.Close()
	for {
		message, ok := <-clientInstance.Send
		if !ok {
			clientInstance.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		clientInstance.Conn.WriteJSON(message)
	}
}

func (hubInstance *Hub) AddClient(client *Client) {
	hubInstance.mutex.Lock()
	defer hubInstance.mutex.Unlock()
	if hubInstance.connections[client.UserID] == nil {
		hubInstance.connections[client.UserID] = make(map[*Client]bool)
	}
	hubInstance.connections[client.UserID][client] = true
}

func (hubInstance *Hub) RemoveClient(client *Client) {
	hubInstance.mutex.Lock()
	defer hubInstance.mutex.Unlock()
	if connectionsMap, ok := hubInstance.connections[client.UserID]; ok {
		if _, exists := connectionsMap[client]; exists {
			delete(connectionsMap, client)
			close(client.Send)
			if len(connectionsMap) == 0 {
				delete(hubInstance.connections, client.UserID)
			}
		}
	}
}

func (hubInstance *Hub) SendToUser(userID int, message interface{}) {
	hubInstance.mutex.RLock()
	defer hubInstance.mutex.RUnlock()
	if connectionsMap, ok := hubInstance.connections[userID]; ok {
		for client := range connectionsMap {
			select {
			case client.Send <- message:
			default:
			}
		}
	}
}

func (hubInstance *Hub) Broadcast(message interface{}) {
	hubInstance.mutex.RLock()
	defer hubInstance.mutex.RUnlock()
	for _, connectionsMap := range hubInstance.connections {
		for client := range connectionsMap {
			select {
			case client.Send <- message:
			default:
			}
		}
	}
}
