package handlers

import (
	"coffeebase-api/internal/middleware"
	"coffeebase-api/internal/notifications"
	"net/http"

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

	// Keep connection alive until client disconnects
	// The read loop runs in the main handler goroutine
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
