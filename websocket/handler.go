package websocket

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kiRiLL3311/Go-multi-chat/middleware"
)

// Client struct to hold connection and user info
type Client struct {
	Conn     *websocket.Conn
	Username string
}

var (
	wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	// The client list now maps connections to Client structs
	clients = make(map[*websocket.Conn]*Client)
	mutex   = &sync.Mutex{}
)

func HandleConnections(c *gin.Context) {
	user, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.Status(http.StatusUnauthorized)
		slog.Error("Ws: User not in context")
		return
	}

	// Upgrade the HTTP connection to a WebSocket connection
	ws, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("Ws: Websocket upgrade error", "error", err)
		return
	}
	defer ws.Close()

	// Register the new client
	client := &Client{Conn: ws, Username: user.Username}
	mutex.Lock()
	clients[ws] = client
	mutex.Unlock()

	slog.Info("Ws: Client connected", "username", client.Username)
	broadcast(fmt.Sprintf("System: %s joined the chat", client.Username))

	// Handle incoming messages
	for {
		_, messageBytes, err := ws.ReadMessage()
		if err != nil {
			// disconnection
			break
		}
		// Prepend the username to the message
		fullMessage := fmt.Sprintf("%s: %s", client.Username, string(messageBytes))
		broadcast(fullMessage)
	}

	// Unregister the client on disconnection
	mutex.Lock()
	delete(clients, ws)
	mutex.Unlock()

	broadcast(fmt.Sprintf("System: %s left the chat", client.Username))
	slog.Info("Ws: Client disconnected", "username", client.Username)
}

// Broadcast function to send to all connected clients
func broadcast(message string) {
	mutex.Lock()
	defer mutex.Unlock()

	slog.Info("Ws: Broadcasting message", "message", message)
	for conn := range clients {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			slog.Info("Ws: Error broadcasting to client", "error", err)
			conn.Close()
			delete(clients, conn)
		}
	}
}
