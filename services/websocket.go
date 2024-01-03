package services

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var clients = make(map[string]*Client) // map of clients
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // Adjust this for your needs
    },
}

type Client struct {
    userID   string
    conn     *websocket.Conn
    send     chan []byte
}
type Message struct {
    ChatID    string    `json:"chat_id"`
    SenderID    string `json:"sender_id"`
    RecipientID string `json:"recipient_id"`
    Content   string `json:"content"`
}

func ServeWs(c *gin.Context) {
    userID := c.Query("userId") 
    log.Printf("Attempting to serve websocket for user %s", userID)
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket upgrade failed: " + err.Error()})
        return
    }
    client := &Client{userID: userID, conn: conn, send: make(chan []byte)}
    clients[userID] = client
    log.Printf("User %s connected. Total clients: %d", userID, len(clients))
    log.Printf("Current clients: %+v", clients)


    go client.writePump()
    go client.readPump()
}

func (c *Client) readPump() {
    defer func() {
        log.Printf("Closing connection for user %s", c.userID)

        c.conn.Close()
        delete(clients, c.userID)
        log.Printf("User %s disconnected. Total clients: %d", c.userID, len(clients))
    }()
    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
        log.Println("Received message:", string(message)) // Log the received message for debugging

        var msg Message
        err = json.Unmarshal(message, &msg)
        if err != nil {
            log.Printf("error: %v", err)
            continue
        }
        
        if recipient, ok := clients[msg.RecipientID]; ok {
            recipient.send <- message
            log.Printf("Routing message from %s to %s", msg.SenderID, msg.RecipientID)
        } else {
            log.Printf("No active client with ID %s", msg.RecipientID)
        }
    }
}

func (c *Client) writePump() {
    defer c.conn.Close()
    for {
        message, ok := <-c.send
        if !ok {
            log.Printf("Channel closed for user %s", c.userID)

            c.conn.WriteMessage(websocket.CloseMessage, []byte{})
            break
        }
        log.Printf("Sending message to user %s: %s", c.userID, string(message))

        c.conn.WriteMessage(websocket.TextMessage, message)
    }
}
