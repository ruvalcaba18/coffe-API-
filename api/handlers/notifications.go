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
	notificationHub *notifications.Hub
}

// --- Public ---

func NewNotificationHandler(notificationHub *notifications.Hub) *NotificationHandler {
	return &NotificationHandler{
		notificationHub: notificationHub,
	}
}

func (notificationHandler *NotificationHandler) HandleWS(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	currentUserID := httpRequest.Context().Value(middleware.UserIDKey).(int)

	websocketConnection, upgradeError := websocketUpgrader.Upgrade(responseWriter, httpRequest, nil)
	if upgradeError != nil {
		return
	}
	defer websocketConnection.Close()

	notificationClient := &notifications.Client{
		Hub:    notificationHandler.notificationHub,
		UserID: currentUserID,
		Conn:   websocketConnection,
		Send:   make(chan interface{}, 256),
	}

	notificationHandler.notificationHub.AddClient(notificationClient)
	defer notificationHandler.notificationHub.RemoveClient(notificationClient)

	go notificationClient.WritePump()

	for {
		var incomingMessage map[string]interface{}
		readingError := websocketConnection.ReadJSON(&incomingMessage)
		if readingError != nil {
			break
		}

		notificationHandler.handleIncomingWSMessage(currentUserID, incomingMessage)
	}
}

// --- Private ---

func (notificationHandler *NotificationHandler) handleIncomingWSMessage(userID int, message map[string]interface{}) {
	if message["type"] == "chat_message" {
		go func() {
			time.Sleep(1 * time.Second)
			autoReply := map[string]interface{}{
				"type":    "chat_message",
				"message": "¡Recibido! Un barista se pondrá en contacto contigo pronto.",
				"sender":  "support",
			}
			notificationHandler.notificationHub.SendToUser(userID, autoReply)
		}()
	}
}
