package handler

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"go-forum-project/chat-service/internal/middleware"
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
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msgReq MessageRequest
		if err := json.Unmarshal(message, &msgReq); err != nil {
			log.Printf("error unmarshaling message request: %v", err)
			continue
		}

		switch msgReq.Action {
		case "create":
			var payload CreateMessagePayload
			if err := json.Unmarshal(msgReq.Payload, &payload); err != nil {
				log.Printf("error unmarshaling message payload: %v", err)
				continue
			}

			if err := c.hub.useCase.CreateMessage(context.Background(), payload.Author, payload.Text); err != nil {
				log.Printf("error creating message: %v", err)
				continue
			}

			c.hub.broadcastMessages()

		case "delete":
			var payload DeleteMessagePayload
			if err := json.Unmarshal(msgReq.Payload, &payload); err != nil {
				log.Printf("error unmarshaling delete payload: %v", err)
				continue
			}

			if err := c.hub.useCase.DeleteMessage(context.Background(), payload.ID); err != nil {
				log.Printf("error deleting message: %v", err)
				continue
			}

			c.hub.broadcastMessages()

		case "get_all":
			c.sendMessages()

		default:
			log.Printf("unknown action: %s", msgReq.Action)
		}
	}
}

func (h *Hub) broadcastMessages() {
	messages, err := h.useCase.GetAllMessages(context.Background())
	if err != nil {
		log.Printf("error getting messages for broadcast: %v", err)
		return
	}

	msgBytes, err := json.Marshal(messages)
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

func ServeWs(hub *Hub, authMiddleware *middleware.WebSocketAuth) http.HandlerFunc {
	return authMiddleware.Middleware(func(w http.ResponseWriter, r *http.Request) {
		conn, err := hub.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket upgrade error:", err)
			return
		}

		username, ok := r.Context().Value("username").(string)
		if !ok || username == "" {
			conn.WriteMessage(websocket.CloseMessage, []byte("Unauthorized"))
			conn.Close()
			return
		}

		client := &Client{
			hub:      hub,
			conn:     conn,
			send:     make(chan []byte, 256),
			username: username,
		}

		client.hub.register <- client

		go client.writePump()
		go client.readPump()
	})
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
