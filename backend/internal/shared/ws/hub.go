package ws

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"

	"monorepo/backend/internal/shared/jwt"
	"monorepo/backend/internal/shared/response"
)

const (
	EventTicketCreated = "ticket.created"
	EventTicketUpdated = "ticket.updated"
)

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type Client struct {
	userID string
	role   string
	send   chan []byte
}

func NewClient(userID, role string) *Client {
	return &Client{
		userID: strings.TrimSpace(userID),
		role:   strings.ToLower(strings.TrimSpace(role)),
		send:   make(chan []byte, 32),
	}
}

func (c *Client) SendChannel() <-chan []byte {
	return c.send
}

type Hub struct {
	mu        sync.RWMutex
	clients   map[*Client]struct{}
	jwtSecret string
}

func NewHub(jwtSecret string) *Hub {
	return &Hub{
		clients:   make(map[*Client]struct{}),
		jwtSecret: jwtSecret,
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client] = struct{}{}
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
	}
}

func ShouldReceive(userID, role, requesterID string) bool {
	normalizedRole := strings.ToLower(strings.TrimSpace(role))
	if normalizedRole == "admin" || normalizedRole == "agent" {
		return true
	}
	return normalizedRole == "requester" && strings.TrimSpace(userID) != "" && userID == requesterID
}

func (h *Hub) BroadcastTicketEvent(eventType string, payload any, requesterID string) {
	raw, err := json.Marshal(Event{
		Type: eventType,
		Data: payload,
	})
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if !ShouldReceive(client.userID, client.role, requesterID) {
			continue
		}
		select {
		case client.send <- raw:
		default:
			// Drop slow client message rather than blocking the broadcast loop.
		}
	}
}

func (h *Hub) RegisterRoutes(router fiber.Router) {
	wsGroup := router.Group("/ws")

	wsGroup.Use("/tickets", func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		token := strings.TrimSpace(c.Query("token"))
		if token == "" {
			return response.Unauthorized(c, "missing token")
		}

		claims, err := jwt.ParseToken(h.jwtSecret, token)
		if err != nil {
			return response.Unauthorized(c, "invalid token")
		}

		c.Locals("ws_user_id", claims.Subject)
		c.Locals("ws_user_role", claims.Role)
		return c.Next()
	})

	wsGroup.Get("/tickets", websocket.New(func(conn *websocket.Conn) {
		userID, _ := conn.Locals("ws_user_id").(string)
		role, _ := conn.Locals("ws_user_role").(string)

		client := NewClient(userID, role)
		h.Register(client)

		done := make(chan struct{})
		go func() {
			defer close(done)
			for msg := range client.send {
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}
			}
		}()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}

		h.Unregister(client)
		_ = conn.Close()
		<-done
	}))
}
