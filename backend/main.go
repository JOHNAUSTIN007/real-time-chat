package main

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

var (
	// Initialize Redis client
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis address
	})
	ctx = context.Background()
)

// Message represents a chat message
type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

// Client represents a connected client
type Client struct {
	conn *websocket.Conn
	send chan Message
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			// Save message to Redis (optional)
			// For simplicity, we'll skip persistence
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

var hub = newHub()

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket Upgrade error:", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan Message),
	}

	hub.register <- client

	// Start goroutines for reading and writing
	go client.readPump()
	go client.writePump()
}

func (c *Client) readPump() {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}
		hub.broadcast <- msg
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for msg := range c.send {
		err := c.conn.WriteJSON(msg)
		if err != nil {
			log.Println("WebSocket write error:", err)
			break
		}
	}
}

func main() {
	// Test Redis connection
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Could not connect to Redis:", err)
	}

	// Start the hub
	go hub.run()

	router := gin.Default()

	router.GET("/ws", handleWebSocket)

	// Simple health check
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	router.Run(":8080") // Run on port 8080
}
