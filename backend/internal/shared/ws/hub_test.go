package ws_test

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	fastws "github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"

	"monorepo/backend/internal/shared/jwt"
	"monorepo/backend/internal/shared/ws"
)

func TestShouldReceive(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		role        string
		requesterID string
		want        bool
	}{
		{name: "admin receives all", userID: "admin-1", role: "admin", requesterID: "req-1", want: true},
		{name: "agent receives all", userID: "agent-1", role: "agent", requesterID: "req-1", want: true},
		{name: "matching requester receives own", userID: "req-1", role: "requester", requesterID: "req-1", want: true},
		{name: "other requester denied", userID: "req-2", role: "requester", requesterID: "req-1", want: false},
		{name: "unknown role denied", userID: "u-1", role: "guest", requesterID: "u-1", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ws.ShouldReceive(tc.userID, tc.role, tc.requesterID); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestHub_BroadcastTicketEvent(t *testing.T) {
	hub := ws.NewHub("secret")
	adminClient := ws.NewClient("admin-1", "admin")
	req1Client := ws.NewClient("req-1", "requester")
	req2Client := ws.NewClient("req-2", "requester")

	hub.Register(adminClient)
	hub.Register(req1Client)
	hub.Register(req2Client)
	defer hub.Unregister(adminClient)
	defer hub.Unregister(req1Client)
	defer hub.Unregister(req2Client)

	hub.BroadcastTicketEvent(ws.EventTicketCreated, map[string]string{"id": "t-1"}, "req-1")

	select {
	case msg := <-adminClient.SendChannel():
		var ev ws.Event
		if err := json.Unmarshal(msg, &ev); err != nil {
			t.Fatalf("unmarshal admin event: %v", err)
		}
		if ev.Type != ws.EventTicketCreated {
			t.Fatalf("expected %s, got %s", ws.EventTicketCreated, ev.Type)
		}
	default:
		t.Fatal("expected admin client to receive broadcast")
	}

	select {
	case <-req1Client.SendChannel():
	default:
		t.Fatal("expected matching requester client to receive broadcast")
	}

	select {
	case msg := <-req2Client.SendChannel():
		t.Fatalf("expected non-matching requester not to receive broadcast, got %s", string(msg))
	default:
	}
}

func TestHub_WebSocketEndToEnd(t *testing.T) {
	const secret = "ws-test-secret"
	hub := ws.NewHub(secret)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	api := app.Group("/api/v1")
	hub.RegisterRoutes(api)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		_ = app.Listener(ln)
	}()
	defer func() { _ = app.Shutdown() }()

	adminToken, err := jwt.GenerateToken(secret, time.Minute, "admin-1", "admin@example.com", "Admin", "admin")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	wsURL := "ws://" + ln.Addr().String() + "/api/v1/ws/tickets?token=" + adminToken
	conn, _, err := fastws.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	// Give the server goroutine a moment to register the client.
	time.Sleep(50 * time.Millisecond)

	hub.BroadcastTicketEvent(ws.EventTicketCreated, map[string]string{
		"id":    "t-99",
		"title": "Live Ticket",
	}, "req-1")

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read websocket message: %v", err)
	}

	var ev ws.Event
	if err := json.Unmarshal(raw, &ev); err != nil {
		t.Fatalf("unmarshal ws event: %v", err)
	}
	if ev.Type != ws.EventTicketCreated {
		t.Fatalf("expected event %s, got %s", ws.EventTicketCreated, ev.Type)
	}
}
