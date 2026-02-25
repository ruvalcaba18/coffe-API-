package handlers

import (
	"coffeebase-api/internal/middleware"
	"coffeebase-api/internal/notifications"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, check against whitelist
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

	h.Hub.AddConnection(userID, conn)
	defer h.Hub.RemoveConnection(userID, conn)

	// Keep connection alive until client disconnects
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
