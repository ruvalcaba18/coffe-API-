package handlers

import (
	"coffeebase-api/internal/middleware"
	"coffeebase-api/internal/notifications"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var websocketUpgrader = websocket.Upgrader{
	CheckOrigin: func(request *http.Request) bool {
		return true
	},
}

type NotificationHandler struct {
	NotificationHub *notifications.Hub
}

func (notificationHandler *NotificationHandler) HandleWS(responseWriter http.ResponseWriter, request *http.Request) {
	currentUserID := request.Context().Value(middleware.UserIDKey).(int)

	websocketConnection, upgradeError := websocketUpgrader.Upgrade(responseWriter, request, nil)
	if upgradeError != nil {
		return
	}
	defer websocketConnection.Close()

	notificationClient := &notifications.Client{
		Hub:    notificationHandler.NotificationHub,
		UserID: currentUserID,
		Conn:   websocketConnection,
		Send:   make(chan interface{}, 256),
	}

	notificationHandler.NotificationHub.AddClient(notificationClient)
	defer notificationHandler.NotificationHub.RemoveClient(notificationClient)

	// Iniciar la goroutine de escritura asíncrona
	go notificationClient.WritePump()

	// Keep connection alive and read incoming messages
	for {
		var incomingMessage map[string]interface{}
		readingError := websocketConnection.ReadJSON(&incomingMessage)
		if readingError != nil {
			break
		}

		// Handle Incoming Message
		if incomingMessage["type"] == "chat_message" {
			// Broadcast to all (for now) or process
			// Example: Auto-reply from Barista
			go func() {
				time.Sleep(1 * time.Second)
				autoReplyMessage := map[string]interface{}{
					"type":    "chat_message",
					"message": "¡Recibido! Un barista se pondrá en contacto contigo pronto.",
					"sender":  "support",
				}
				notificationHandler.NotificationHub.SendToUser(currentUserID, autoReplyMessage)
			}()
		}
	}
}
