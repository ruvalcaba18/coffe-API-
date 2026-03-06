package handlers

import (
	"coffeebase-api/internal/middleware"
	"coffeebase-api/internal/notifications"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type NotificationHandler struct {
	Hub *notifications.Hub
}

func (h *NotificationHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	client := &notifications.Client{
		Hub:    h.Hub,
		UserID: userID,
		Conn:   conn,
		Send:   make(chan interface{}, 256),
	}

	h.Hub.AddClient(client)
	defer h.Hub.RemoveClient(client)

	// Iniciar la goroutine de escritura asíncrona
	go client.WritePump()

	// Keep connection alive and read incoming messages
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		// Handle Incoming Message
		if msg["type"] == "chat_message" {
			// Broadcast to all (for now) or process
			// Example: Auto-reply from Barista
			go func() {
				time.Sleep(1 * time.Second)
				reply := map[string]interface{}{
					"type":    "chat_message",
					"message": "¡Recibido! Un barista se pondrá en contacto contigo pronto.",
					"sender":  "support",
				}
				h.Hub.SendToUser(userID, reply)
			}()
		}
	}
}
