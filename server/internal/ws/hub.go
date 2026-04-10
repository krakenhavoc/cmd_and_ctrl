// Package ws wraps gorilla/websocket with a small hub-and-client pattern
// following the canonical gorilla chat example. At v0 the hub just holds
// connected clients and dispatches ping frames. S02 and S03 grow this into
// per-game rooms and an action router.
package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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

	// room is the single S03 default game. If nil, the hub accepts
	// ping/pong only and rejects action frames with an internal
	// error. S04 will replace the singleton with a RoomManager.
	//
	// Stored in an atomic.Pointer so SetRoom and the read-side accesses
	// in ServeWS/handleAction don't race, even though the current
	// production call pattern (SetRoom on startup, before the HTTP
	// server starts) is serialised by construction.
	room atomic.Pointer[Room]
}

// NewHub creates an empty hub with no room attached.
func NewHub(log *slog.Logger) *Hub {
	if log == nil {
		log = slog.Default()
	}
	return &Hub{
		clients: make(map[*Client]struct{}),
		log:     log,
	}
}

// SetRoom attaches a room to the hub. The atomic pointer means this
// is race-free against concurrent reads, but the current production
// call pattern (SetRoom on startup, before the HTTP server starts)
// also serialises it by construction.
func (h *Hub) SetRoom(r *Room) {
	h.room.Store(r)
}

// loadRoom returns the currently attached room, or nil if none.
func (h *Hub) loadRoom() *Room {
	return h.room.Load()
}

// admit atomically adds a client to the hub AND increments the pump
// WaitGroup by 2, under a single hold of h.mu. Returns false if the
// hub has been shut down, in which case the caller must close the
// connection and bail without spawning pumps.
//
// The combined register + wg.Add invariant is load-bearing: if wg.Add
// ran outside the lock, Shutdown could observe a client in h.clients,
// call wg.Wait() (which returns immediately if no pump has Added yet
// for this client), and then the caller's subsequent wg.Add would be
// "Add after Wait with counter 0" — documented sync.WaitGroup misuse
// that panics at runtime.
func (h *Hub) admit(c *Client) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return false
	}
	h.clients[c] = struct{}{}
	h.wg.Add(2)
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

	// Pre-stage the initial snapshot into the client's send channel
	// BEFORE the client becomes visible to broadcastAll. This closes
	// the "initial snapshot arrives after a concurrent action's
	// broadcast" race: since the client is not yet in h.clients, no
	// concurrent Apply can broadcast to this channel, so the initial
	// snapshot is guaranteed to be the first frame the new client
	// sees. The channel is freshly-created and buffered at 16, so the
	// non-blocking send inside sendRaw always succeeds here.
	room := h.loadRoom()
	if room != nil {
		if raw, seq, err := room.Snapshot(); err == nil {
			client.sendRaw(raw)
			client.log.Debug("pre-staged initial snapshot", "seq", seq)
		} else {
			client.log.Error("build initial snapshot failed", "err", err)
		}
	}

	// admit() atomically registers the client AND increments the wg
	// by 2 under a single hold of h.mu. That combined critical section
	// prevents Shutdown from observing a client whose wg.Add hasn't
	// run yet, which would be "Add after Wait with counter 0" — a
	// runtime panic.
	if !h.admit(client) {
		// A Shutdown raced us between the fast-path check and admit.
		// Close the freshly-upgraded socket and bail. No pumps have
		// been spawned, so no wg.Done is owed.
		_ = conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutting down"),
			time.Now().Add(writeWait),
		)
		_ = conn.Close()
		return
	}
	client.log.Info("ws client connected", "total", h.Count())

	go func() {
		defer h.wg.Done()
		client.writePump()
	}()
	go func() {
		defer h.wg.Done()
		client.readPump()
	}()
}

// broadcastAll sends a pre-marshalled frame to every currently
// connected client. The hub's read lock prevents unregister from
// closing any client's send channel mid-broadcast, so the sends are
// always to open channels. A full send buffer disconnects the slow
// client rather than dropping the frame silently — see sendFrame for
// the same policy.
func (h *Hub) broadcastAll(raw []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		c.sendRaw(raw)
	}
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
	case protocol.KindAction:
		c.handleAction(frame)
	default:
		c.sendError(frame.ID, protocol.CodeBadRequest,
			"unknown or unsupported kind")
	}
}

// handleAction decodes an action payload, dispatches it to the game
// via the actions package, and broadcasts a fresh snapshot to every
// connected client on success. Validation or dispatch errors come
// back as error frames to the originating client only.
//
// Dispatch + seq alloc + state capture are serialised under the
// room's mutex via Room.Apply, so two concurrent action frames cannot
// interleave their state changes or broadcast non-monotonic snapshots.
func (c *Client) handleAction(frame protocol.Frame) {
	room := c.hub.loadRoom()
	if room == nil {
		// Server configuration bug, not a client mistake.
		c.sendError(frame.ID, protocol.CodeInternal, "server has no room configured")
		return
	}

	// An explicit JSON `null` payload has len > 0 but unmarshals into
	// the zero-value ActionPayload, which would then fall through the
	// Dispatch switch with an empty Type and produce a cryptic
	// "unknown action type" error. Treat `null` the same as an absent
	// payload: it's a malformed request.
	if len(frame.Payload) == 0 || bytes.Equal(bytes.TrimSpace(frame.Payload), []byte("null")) {
		c.sendError(frame.ID, protocol.CodeBadRequest, "action frame missing payload")
		return
	}
	var payload protocol.ActionPayload
	if err := json.Unmarshal(frame.Payload, &payload); err != nil {
		c.sendError(frame.ID, protocol.CodeBadJSON, "action payload is not valid JSON")
		return
	}

	action, err := actions.Decode(payload.Type, payload.Player, payload.Params)
	if err != nil {
		c.sendError(frame.ID, protocol.CodeBadRequest, err.Error())
		return
	}

	raw, seq, err := room.Apply(func() error {
		return actions.Dispatch(room.Game, action)
	})
	if err != nil {
		code, msg := classifyActionError(err)
		c.sendError(frame.ID, code, msg)
		return
	}
	c.hub.broadcastAll(raw)
	c.log.Debug("action dispatched", "type", payload.Type, "seq", seq)
}

// classifyActionError maps a dispatch-time error from the game or
// actions package to a wire error code + clean client-facing message.
// Internal package prefixes (`game:`, `actions:`) are stripped from
// the message so the wire surface doesn't leak Go module layout, and
// sentinel errors that reflect server state problems get the
// `internal` code instead of `bad_request`.
func classifyActionError(err error) (code, message string) {
	if err == nil {
		return "", ""
	}
	// Server-state errors: the client's request is well-formed, but
	// the game is in a state that can't satisfy it. Report as internal
	// so clients don't treat it as "your input was malformed".
	switch {
	case errors.Is(err, game.ErrGameNotActive):
		return protocol.CodeInternal, "game is not in active state"
	case errors.Is(err, game.ErrZoneEmpty):
		// e.g. draw_card on an empty library — the player has lost
		// (state-based action) but S03 doesn't enforce that, so the
		// action is well-formed but unsatisfiable.
		return protocol.CodeInternal, "zone is empty"
	}
	// Everything else is traceable to a client-supplied input — bad
	// player ID, bad card ID, bad zone, unknown action type, etc.
	msg := err.Error()
	msg = strings.TrimPrefix(msg, "game: ")
	msg = strings.TrimPrefix(msg, "actions: ")
	return protocol.CodeBadRequest, msg
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
	c.sendRaw(raw)
}

// sendRaw enqueues a pre-marshalled frame on the client's send
// channel. A full buffer disconnects the client rather than dropping
// the frame silently — letting the pong for a ping, or the snapshot
// for an action, be quietly dropped would mask real bugs.
func (c *Client) sendRaw(raw []byte) {
	select {
	case c.send <- raw:
	default:
		c.log.Warn("ws send buffer full, disconnecting client")
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
//
// The per-client WriteControl + Close loop honours ctx: if the caller's
// deadline elapses mid-loop we stop cleanly rather than burning through
// writeWait-many seconds on every dead client.
func (h *Hub) Shutdown(ctx context.Context) {
	h.mu.Lock()
	h.closed = true
	clients := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.Unlock()

	for _, c := range clients {
		if ctx.Err() != nil {
			break
		}
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
