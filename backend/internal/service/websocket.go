package service

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"github.com/instaagrammeta/somon-crm/backend/internal/models"
)

// Topic strings used in Event.Topic. Frontend subscribes by topic prefix:
//
//	notif      → all notifications for the current user
//	chat       → chat messages, optionally suffixed by room: chat:<peer_id>
//	kanban     → kanban board changes (board id can be appended)
//	presence   → online/offline list updates
//	tasks      → task created/updated/deleted (subscribed by all on /zadacha)
//	leads      → kanban-leads moves (subscribed by all on /baza)
//	ai         → AI long-running job completions
const (
	TopicNotif    = "notif"
	TopicChat     = "chat"
	TopicKanban   = "kanban"
	TopicPresence = "presence"
	TopicTasks    = "tasks"
	TopicLeads    = "leads"
	TopicAI       = "ai"
)

// Event is the wire format for WS messages (both directions).
type Event struct {
	Topic   string `json:"topic"`
	Action  string `json:"action,omitempty"` // create|update|delete|ping|read|typing
	From    uint   `json:"from,omitempty"`   // sender user id
	To      uint   `json:"to,omitempty"`     // recipient user id (chat DM)
	Payload any    `json:"payload,omitempty"`
	At      int64  `json:"at,omitempty"` // unix ms
}

// client represents a single WebSocket connection.
type client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID uint
	role   string
}

// Hub fans out events to connected clients. One hub per server process; for
// multi-replica deployments swap the in-memory map for a Redis Pub/Sub layer.
type Hub struct {
	db    *gorm.DB
	mu    sync.RWMutex
	users map[uint]map[*client]struct{} // userID → set of connections
	all   map[*client]struct{}
}

func NewHub(db *gorm.DB) *Hub {
	return &Hub{
		db:    db,
		users: make(map[uint]map[*client]struct{}),
		all:   make(map[*client]struct{}),
	}
}

// upgrader is shared. CheckOrigin returns true because the origin is already
// validated by the CORS middleware before the handshake reaches us.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Serve upgrades the HTTP request to a WS connection and handles the lifetime.
// userID is taken from the auth middleware (after JWT verification).
func (h *Hub) Serve(w http.ResponseWriter, r *http.Request, userID uint, role string) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	c := &client{hub: h, conn: conn, send: make(chan []byte, 64), userID: userID, role: role}
	h.register(c)

	go c.writeLoop()
	go c.readLoop()
	return nil
}

func (h *Hub) register(c *client) {
	h.mu.Lock()
	if h.users[c.userID] == nil {
		h.users[c.userID] = make(map[*client]struct{})
	}
	h.users[c.userID][c] = struct{}{}
	h.all[c] = struct{}{}
	h.mu.Unlock()
	h.markPresence(c.userID, true)
	h.broadcastPresence(c.userID, true)
}

func (h *Hub) unregister(c *client) {
	h.mu.Lock()
	delete(h.all, c)
	if conns, ok := h.users[c.userID]; ok {
		delete(conns, c)
		if len(conns) == 0 {
			delete(h.users, c.userID)
			h.mu.Unlock()
			h.markPresence(c.userID, false)
			h.broadcastPresence(c.userID, false)
			close(c.send)
			_ = c.conn.Close()
			return
		}
	}
	h.mu.Unlock()
	close(c.send)
	_ = c.conn.Close()
}

func (h *Hub) markPresence(userID uint, online bool) {
	if h.db == nil {
		return
	}
	row := models.UserPresence{UserID: userID, LastSeen: time.Now(), Online: online}
	h.db.Save(&row)
}

func (h *Hub) broadcastPresence(userID uint, online bool) {
	h.Broadcast(Event{
		Topic:   TopicPresence,
		Action:  "update",
		From:    userID,
		Payload: map[string]any{"user_id": userID, "online": online},
		At:      time.Now().UnixMilli(),
	})
}

// Broadcast sends an event to every connected client. The event's audience can
// be narrowed by Topic conventions on the frontend (it filters by topic
// prefix), so we don't shard rooms server-side.
func (h *Hub) Broadcast(ev Event) {
	if ev.At == 0 {
		ev.At = time.Now().UnixMilli()
	}
	raw, err := json.Marshal(ev)
	if err != nil {
		return
	}
	h.mu.RLock()
	for c := range h.all {
		select {
		case c.send <- raw:
		default:
			// Slow consumer — drop this message rather than blocking the hub.
		}
	}
	h.mu.RUnlock()
}

// SendToUser delivers an event to every connection owned by userID.
func (h *Hub) SendToUser(userID uint, ev Event) {
	if ev.At == 0 {
		ev.At = time.Now().UnixMilli()
	}
	raw, _ := json.Marshal(ev)
	h.mu.RLock()
	conns := h.users[userID]
	h.mu.RUnlock()
	for c := range conns {
		select {
		case c.send <- raw:
		default:
		}
	}
}

// SendToUsers fans an event out to a small list of recipients.
func (h *Hub) SendToUsers(userIDs []uint, ev Event) {
	for _, uid := range userIDs {
		h.SendToUser(uid, ev)
	}
}

// OnlineUsers returns the set of currently-connected user IDs.
func (h *Hub) OnlineUsers() []uint {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]uint, 0, len(h.users))
	for uid := range h.users {
		out = append(out, uid)
	}
	return out
}

// ===== client read/write loops =====

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 64 * 1024
)

func (c *client) readLoop() {
	defer c.hub.unregister(c)
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		var ev Event
		if err := json.Unmarshal(msg, &ev); err != nil {
			continue
		}
		ev.From = c.userID
		ev.At = time.Now().UnixMilli()
		// Echo "typing" / "read" to the recipient only; everything else is
		// dropped (clients should write through REST, not WS, for persistent
		// data — WS is only an event channel).
		switch ev.Action {
		case "typing", "read", "ping":
			if ev.To != 0 {
				c.hub.SendToUser(ev.To, ev)
			} else {
				c.hub.Broadcast(ev)
			}
		}
	}
}

func (c *client) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
