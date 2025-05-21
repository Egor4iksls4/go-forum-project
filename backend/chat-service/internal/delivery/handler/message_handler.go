package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"go-forum-project/chat-service/internal/client"
	"go-forum-project/chat-service/internal/usecase"
	"log"
	"net/http"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	username string
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	useCase    usecase.MessageUseCase
	upgrader   *websocket.Upgrader
}

func NewHub(uc usecase.MessageUseCase) *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		useCase:    uc,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *Hub) Run() {
	go h.cleanupOldMessages()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

type MessageRequest struct {
	Action  string          `json:"action"` // "create", "delete", "get_all"
	Payload json.RawMessage `json:"payload"`
}

// CreateMessagePayload структура для создания сообщения
type CreateMessagePayload struct {
	Author string `json:"author"`
	Text   string `json:"text"`
}

// DeleteMessagePayload структура для удаления сообщения
type DeleteMessagePayload struct {
	ID int `json:"id"`
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, rawMsg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Логируем сырое сообщение для отладки
		log.Printf("Raw message received: %s", string(rawMsg))

		// Обработка сообщения
		if err := c.handleMessage(rawMsg); err != nil {
			log.Printf("Error handling message: %v", err)
		}
	}
}

func (c *Client) handleMessage(rawMsg []byte) error {
	// Базовый парсинг для определения типа сообщения
	var baseMsg struct {
		Action  string          `json:"action"`
		Payload json.RawMessage `json:"payload"`
	}

	if err := json.Unmarshal(rawMsg, &baseMsg); err != nil {
		return fmt.Errorf("invalid message format: %v", err)
	}

	switch baseMsg.Action {
	case "create":
		return c.handleCreateMessage(baseMsg.Payload)
	case "delete":
		return c.handleDeleteMessage(baseMsg.Payload)
	case "get_all":
		c.sendMessages()
		return nil
	default:
		return fmt.Errorf("unknown action: %s", baseMsg.Action)
	}
}

func (c *Client) handleCreateMessage(payload json.RawMessage) error {
	var createMsg struct {
		Author string `json:"author"`
		Text   string `json:"text"`
	}

	if err := json.Unmarshal(payload, &createMsg); err != nil {
		return fmt.Errorf("invalid create message payload: %v", err)
	}

	if createMsg.Text == "" {
		return errors.New("empty message text")
	}

	if err := c.hub.useCase.CreateMessage(context.Background(), createMsg.Author, createMsg.Text); err != nil {
		return fmt.Errorf("error creating message: %v", err)
	}

	c.hub.broadcastMessages()
	return nil
}

func (c *Client) handleDeleteMessage(payload json.RawMessage) error {
	var deleteMsg struct {
		ID int `json:"id"`
	}

	if err := json.Unmarshal(payload, &deleteMsg); err != nil {
		return fmt.Errorf("invalid delete message payload: %v", err)
	}

	if err := c.hub.useCase.DeleteMessage(context.Background(), deleteMsg.ID); err != nil {
		return fmt.Errorf("error deleting message: %v", err)
	}

	c.hub.broadcastMessages()
	return nil
}

func (h *Hub) broadcastMessages() {
	messages, err := h.useCase.GetAllMessages(context.Background())
	if err != nil {
		log.Printf("error getting messages for broadcast: %v", err)
		return
	}

	msgToSend := map[string]interface{}{
		"action":  "broadcast",
		"payload": messages,
	}

	msgBytes, err := json.Marshal(msgToSend)
	if err != nil {
		log.Printf("error marshaling messages: %v", err)
		return
	}

	h.broadcast <- msgBytes
}

func (c *Client) sendMessages() {
	messages, err := c.hub.useCase.GetAllMessages(context.Background())
	if err != nil {
		log.Printf("error getting messages: %v", err)
		return
	}

	msgBytes, err := json.Marshal(messages)
	if err != nil {
		log.Printf("error marshaling messages: %v", err)
		return
	}

	c.send <- msgBytes
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWs(hub *Hub, authClient *client.AuthClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.URL.Query().Get("accessToken")

		log.Printf("Access token: %s", accessToken)

		username, valid, err := authClient.ValidateToken(r.Context(), accessToken)
		if err != nil || !valid {
			log.Printf("Token validation error: %v", err)
			respondWithUnauthorized(w, r, hub.upgrader)
			return
		}

		conn, err := hub.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket upgrade error:", err)
			return
		}

		client := &Client{
			hub:      hub,
			conn:     conn,
			send:     make(chan []byte, 256),
			username: username,
		}

		hub.register <- client
		go client.writePump()
		go client.readPump()
	}
}

func respondWithUnauthorized(w http.ResponseWriter, r *http.Request, upgrader *websocket.Upgrader) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(4001, "Unauthorized"))
	conn.Close()
}

func GetMessageHandler(uc usecase.MessageUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		messages, err := uc.GetAllMessages(context.Background())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	}
}

func (h *Hub) cleanupOldMessages() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		if err := h.useCase.CleanupOldMessages(context.Background()); err != nil {
			log.Printf("error cleaning old messages: %v", err)
		} else {
			h.broadcastMessages()
		}
	}
}
