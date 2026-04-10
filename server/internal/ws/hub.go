// Package ws wraps gorilla/websocket with a small hub-and-client pattern
// following the canonical gorilla chat example. At v0 the hub just holds
// connected clients and dispatches ping frames. S02 and S03 grow this into
// per-game rooms and an action router.
package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// Timing constants mirror the canonical gorilla/websocket chat example
// (examples/chat/client.go). Documented inline so future readers don't
// need to find the source.
const (
	writeWait      = 10 * time.Second    // max time to wait for a single outbound write
	pongWait       = 60 * time.Second    // max time to wait for a pong response from a peer
	pingPeriod     = (pongWait * 9) / 10 // ping period must be less than pongWait
	maxMessageSize = 64 * 1024           // max inbound frame size — will grow once game state frames exist
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	// SECURITY: v0 accepts every origin for local dev. This is a
	// cross-site WebSocket hijacking hazard if the server is ever
	// exposed on a public port before S04's shared-password auth lands.
	// Lock this down before the first deploy.
	CheckOrigin: func(_ *http.Request) bool { return true },
}

// Hub is the central registry of connected WebSocket clients. It is safe
// for concurrent use.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
	closed  bool // set by Shutdown; blocks new registrations
	log     *slog.Logger
	wg      sync.WaitGroup // tracks active read/write pumps for graceful shutdown
}

// NewHub creates an empty hub.
func NewHub(log *slog.Logger) *Hub {
	if log == nil {
		log = slog.Default()
	}
	return &Hub{
		clients: make(map[*Client]struct{}),
		log:     log,
	}
}

// register adds a client to the hub. Returns false if the hub has been
// shut down, in which case the caller must not spawn the client pumps.
func (h *Hub) register(c *Client) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return false
	}
	h.clients[c] = struct{}{}
	return true
}

func (h *Hub) unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
	}
}

// Count returns the current number of connected clients.
func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// ServeWS upgrades the HTTP connection to a WebSocket and registers a new
// client. Intended to be mounted at /ws on the HTTP mux. Returns 503 if
// the hub has already been shut down.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Fast path: refuse new connections once shutdown has begun.
	h.mu.RLock()
	closed := h.closed
	h.mu.RUnlock()
	if closed {
		http.Error(w, "server shutting down", http.StatusServiceUnavailable)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("ws upgrade failed", "err", err, "remote", r.RemoteAddr)
		return
	}
	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 16),
		log:  h.log.With("remote", r.RemoteAddr),
	}
	if !h.register(client) {
		// A Shutdown raced us between the fast-path check and
		// register. Close the freshly-upgraded socket and bail.
		_ = conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutting down"),
			time.Now().Add(writeWait),
		)
		_ = conn.Close()
		return
	}
	client.log.Info("ws client connected", "total", h.Count())

	h.wg.Add(2)
	go func() {
		defer h.wg.Done()
		client.writePump()
	}()
	go func() {
		defer h.wg.Done()
		client.readPump()
	}()
}

// Client is one connected WebSocket peer. Each client owns two goroutines:
// a readPump for inbound frames and a writePump for outbound frames. The
// hub never writes to a client's socket directly — it pushes onto the
// client's send channel, and the writePump does the actual socket write.
// This is the canonical gorilla pattern and keeps concurrent writes safe.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
	log  *slog.Logger
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister(c)
		_ = c.conn.Close()
		c.log.Info("ws client disconnected", "total", c.hub.Count())
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log.Warn("ws read error", "err", err)
			}
			return
		}
		c.handleFrame(raw)
	}
}

func (c *Client) handleFrame(raw []byte) {
	var frame protocol.Frame
	if err := json.Unmarshal(raw, &frame); err != nil {
		c.sendError("", protocol.CodeBadJSON, "frame is not valid JSON")
		return
	}
	if frame.V != protocol.Version {
		c.sendError(frame.ID, protocol.CodeBadVersion,
			"unsupported protocol version")
		return
	}

	switch frame.Kind {
	case protocol.KindPing:
		c.handlePing(frame)
	default:
		c.sendError(frame.ID, protocol.CodeBadRequest,
			"unknown or unsupported kind")
	}
}

func (c *Client) handlePing(frame protocol.Frame) {
	var in protocol.PingPayload
	if len(frame.Payload) > 0 {
		if err := json.Unmarshal(frame.Payload, &in); err != nil {
			c.sendError(frame.ID, protocol.CodeBadJSON, "ping payload is not valid JSON")
			return
		}
	}
	out := protocol.PongPayload{
		Msg:        in.Msg,
		ServerTime: time.Now().UTC().Format(time.RFC3339),
	}
	payload, err := json.Marshal(out)
	if err != nil {
		c.sendError(frame.ID, protocol.CodeInternal, "failed to encode pong payload")
		return
	}
	c.sendFrame(protocol.Frame{
		V:       protocol.Version,
		Kind:    protocol.KindPong,
		ID:      frame.ID,
		Payload: payload,
	})
}

func (c *Client) sendFrame(f protocol.Frame) {
	raw, err := json.Marshal(f)
	if err != nil {
		c.log.Error("ws marshal frame", "err", err)
		return
	}
	select {
	case c.send <- raw:
	default:
		// A full send buffer means the writer is stuck or the peer
		// can't keep up. Dropping the frame silently would swallow
		// critical responses (like the pong to a ping). Instead, tear
		// the client down: closing the conn unblocks readPump, which
		// runs the normal unregister + close(send) path.
		c.log.Warn("ws send buffer full, disconnecting client", "kind", f.Kind)
		_ = c.conn.Close()
	}
}

func (c *Client) sendError(id, code, message string) {
	payload, err := json.Marshal(protocol.ErrorPayload{Code: code, Message: message})
	if err != nil {
		c.log.Error("ws marshal error payload", "err", err)
		return
	}
	c.sendFrame(protocol.Frame{
		V:       protocol.Version,
		Kind:    protocol.KindError,
		ID:      id,
		Payload: payload,
	})
}

func (c *Client) writePump() {
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
				c.log.Warn("ws write error", "err", err)
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

// Shutdown marks the hub as closed (so new registrations are rejected),
// closes every connected client, and waits for their read/write pumps
// to exit, or ctx to cancel — whichever comes first. With zero clients,
// this returns essentially immediately. Safe to call once.
func (h *Hub) Shutdown(ctx context.Context) {
	h.mu.Lock()
	h.closed = true
	clients := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.Unlock()

	for _, c := range clients {
		_ = c.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutting down"),
			time.Now().Add(writeWait),
		)
		_ = c.conn.Close()
	}

	done := make(chan struct{})
	go func() {
		h.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
	}
}
