package realtime

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Event struct {
	Type    string      `json:"type"`
	BoardID string      `json:"board_id"`
	Payload interface{} `json:"payload,omitempty"`
}

type Client struct {
	boardID string
	conn    *websocket.Conn
	send    chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]bool
}

func NewHub() *Hub { return &Hub{clients: make(map[string]map[*Client]bool)} }

func (h *Hub) Add(boardID string, conn *websocket.Conn) *Client {
	c := &Client{boardID: boardID, conn: conn, send: make(chan []byte, 32)}
	h.mu.Lock()
	if h.clients[boardID] == nil {
		h.clients[boardID] = make(map[*Client]bool)
	}
	h.clients[boardID][c] = true
	h.mu.Unlock()
	go c.writePump()
	go c.readPump(h)
	return c
}

func (h *Hub) remove(c *Client) {
	h.mu.Lock()
	if room := h.clients[c.boardID]; room != nil {
		if _, ok := room[c]; ok {
			delete(room, c)
			close(c.send)
		}
		if len(room) == 0 {
			delete(h.clients, c.boardID)
		}
	}
	h.mu.Unlock()
	_ = c.conn.Close()
}

func (h *Hub) Broadcast(boardID, eventType string, payload interface{}) {
	data, err := json.Marshal(Event{Type: eventType, BoardID: boardID, Payload: payload})
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[boardID] {
		select {
		case c.send <- data:
		default:
			go h.remove(c)
		}
	}
}

func (c *Client) readPump(h *Hub) {
	defer h.remove(c)
	c.conn.SetReadLimit(1024)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error { return c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)) })
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	defer c.conn.Close()
	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024, WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "" || origin == "http://localhost:5173" || origin == "http://localhost:3000"
	},
}

func (h *Hub) ServeBoard(w http.ResponseWriter, r *http.Request, boardID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade: %v", err)
		return
	}
	h.Add(boardID, conn)
}
