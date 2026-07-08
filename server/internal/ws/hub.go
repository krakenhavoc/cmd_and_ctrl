// Package ws wraps gorilla/websocket with a small hub-and-client pattern
// following the canonical gorilla chat example. As of S04 the hub speaks
// to multiple concurrent games via a RoomManager: every client is bound
// to one game (game UUID in the `?game=` query param) and one optional
// seat (player UUID in `?player=`). Action frames mutate only the bound
// game; broadcasts are scoped to clients sharing that game, each with a
// per-viewer filtered snapshot.
package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
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

// baseUpgrader carries the tunable buffer sizes. The CheckOrigin hook
// is installed per-Hub in NewHub so each hub can honour its own
// allowed-origins configuration without a package-level global.
var baseUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

// Hub is the central registry of connected WebSocket clients. It is safe
// for concurrent use.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
	closed  bool // set by Shutdown; blocks new registrations
	log     *slog.Logger
	wg      sync.WaitGroup // tracks active read/write pumps for graceful shutdown

	// manager holds the set of active Rooms keyed by game ID. A nil
	// manager means the hub accepts only ping/pong and rejects every
	// action frame or multi-game upgrade with an internal error.
	//
	// Stored in an atomic.Pointer so SetManager and the read-side
	// accesses in ServeWS/handleAction don't race. The production
	// call pattern (SetManager on startup) is also serialised by
	// construction, but the atomic makes the ordering explicit.
	manager atomic.Pointer[RoomManager]

	// authorize is an optional hook that runs per WebSocket upgrade
	// request AFTER the request has been admitted at the HTTP layer
	// (status-200 handshake) but BEFORE the connection is registered.
	// It is responsible for extracting the caller's identity (token
	// validation, principal lookup) and enforcing that the caller is
	// allowed to bind to the requested ?game=/?player= pair. A non-nil
	// error rejects the upgrade with an HTTP error; a nil error
	// proceeds with the returned gameID and playerID as the client's
	// bound identity — which may differ from the raw query values
	// (e.g. the authorizer could resolve a session token to a player
	// ID and ignore the query param altogether).
	//
	// If authorize is nil, the hub falls back to reading ?game= and
	// ?player= directly from the query string with no validation —
	// this is the legacy S01–S03 path that pre-lobby tests still use.
	authorize UpgradeAuthorizer

	// upgrader carries this hub's CheckOrigin. Built once in NewHub
	// so hot-path upgrades don't allocate, and so the allowed-origin
	// set can be tweaked via SetAllowedOrigins without touching the
	// package-level state.
	upgrader websocket.Upgrader

	// allowedOrigins is a case-insensitive set of hostnames that
	// CheckOrigin will honour in addition to same-origin. Empty by
	// default (same-origin only). Protected by originsMu.
	originsMu      sync.RWMutex
	allowedOrigins map[string]struct{}
}

// Binding is what the upgrade authorizer hands to the hub: the
// (gameID, playerID) pair the connection should bind to, plus a
// `ReadOnly` flag that gates mutation frames at the hub layer.
//
// ReadOnly = true marks the connection as a spectator (S11) — it
// receives snapshot broadcasts but any incoming `action` frame is
// rejected with bad_request. Admin spectators (admin session, no
// `?player=`) keep ReadOnly = false because admins legitimately
// need to mutate state on a player's behalf.
type Binding struct {
	GameID   uuid.UUID
	PlayerID uuid.UUID
	ReadOnly bool
}

// UpgradeAuthorizer validates an incoming WebSocket upgrade and
// returns the binding the connection should carry. A nil error means
// the upgrade is allowed; any non-nil error is converted to an HTTP
// 401/403 response before the upgrade completes.
//
// The authorizer owns session-token validation, seat-ownership checks,
// and any other policy decisions. It is the *only* place the hub
// looks for identity — the hub itself has no notion of auth.
type UpgradeAuthorizer interface {
	AuthorizeUpgrade(r *http.Request) (Binding, error)
}

// UpgradeAuthorizerFunc is a function adapter for UpgradeAuthorizer,
// handy for tests and for wiring in a one-off closure from main.go
// without declaring a named type.
type UpgradeAuthorizerFunc func(r *http.Request) (Binding, error)

// AuthorizeUpgrade implements UpgradeAuthorizer by calling f.
func (f UpgradeAuthorizerFunc) AuthorizeUpgrade(r *http.Request) (Binding, error) {
	return f(r)
}

// NewHub creates an empty hub with no manager and no authorizer attached.
func NewHub(log *slog.Logger) *Hub {
	if log == nil {
		log = slog.Default()
	}
	h := &Hub{
		clients:        make(map[*Client]struct{}),
		log:            log,
		allowedOrigins: make(map[string]struct{}),
	}
	h.upgrader = baseUpgrader
	h.upgrader.CheckOrigin = h.checkOrigin
	return h
}

// SetAllowedOrigins replaces the set of cross-origin hostnames the hub
// will accept WebSocket upgrades from. Same-origin requests (Origin
// host == Host) and requests with no Origin header (CLI, tests) are
// always allowed. Hostnames are matched case-insensitively; include
// the port (e.g. "192.168.1.20:8080") if it differs from the default.
func (h *Hub) SetAllowedOrigins(origins []string) {
	set := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		// Accept a full URL ("https://lan.local") or a bare host.
		if u, err := url.Parse(o); err == nil && u.Host != "" {
			set[strings.ToLower(u.Host)] = struct{}{}
			continue
		}
		set[strings.ToLower(o)] = struct{}{}
	}
	h.originsMu.Lock()
	h.allowedOrigins = set
	h.originsMu.Unlock()
}

// checkOrigin is the hub's CheckOrigin hook. Same-origin is always
// allowed; cross-origin requires the Origin host to be in the
// configured allow-set. An empty Origin (non-browser clients) is
// treated as same-origin — the browser hijacking threat model only
// applies when a browser is the caller.
func (h *Hub) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		h.log.Warn("ws upgrade rejected: bad Origin", "origin", origin)
		return false
	}
	if strings.EqualFold(u.Host, r.Host) {
		return true
	}
	h.originsMu.RLock()
	_, ok := h.allowedOrigins[strings.ToLower(u.Host)]
	h.originsMu.RUnlock()
	if !ok {
		h.log.Warn("ws upgrade rejected by CheckOrigin", "origin", origin, "host", r.Host)
	}
	return ok
}

// SetManager attaches a RoomManager to the hub. The atomic pointer
// makes this race-free against concurrent reads; the production call
// pattern (SetManager on startup, before the HTTP server starts) also
// serialises it by construction.
func (h *Hub) SetManager(m *RoomManager) {
	h.manager.Store(m)
}

// SetAuthorizer installs a per-upgrade authorization hook. Passing nil
// restores the legacy no-auth path that reads ?game= and ?player=
// from the query string directly. See UpgradeAuthorizer.
func (h *Hub) SetAuthorizer(a UpgradeAuthorizer) {
	h.authorize = a
}

// SetRoom is a backward-compat convenience: it wraps the given room in
// a single-game RoomManager and installs it via SetManager. Tests and
// the single-game dev server (behind CMDCTRL_SEED_DEMO) still call
// this; the multi-game code path uses SetManager directly.
func (h *Hub) SetRoom(r *Room) {
	mgr := NewRoomManager(h.log, "")
	mgr.Register(r)
	h.SetManager(mgr)
}

// loadManager returns the currently attached manager, or nil if none.
func (h *Hub) loadManager() *RoomManager {
	return h.manager.Load()
}

// resolveRoom looks up the room for a connected client. Returns nil
// if the manager has been removed or the client's game ID no longer
// resolves to a live room (e.g. the lobby evicted it mid-session).
func (h *Hub) resolveRoom(c *Client) *Room {
	mgr := h.loadManager()
	if mgr == nil {
		return nil
	}
	return mgr.Get(c.gameID)
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

// ServeWS upgrades the HTTP connection to a WebSocket and registers a
// new client. Intended to be mounted at /ws on the HTTP mux. Returns
// 503 if the hub has already been shut down, 400 if the requested
// game does not resolve, and 401/403 if an authorizer is attached and
// rejects the request.
//
// Query params (when no authorizer is attached):
//
//	?game=<uuid>   — target game ID. Optional if the manager holds
//	                 exactly one room (tests and single-game dev).
//	?player=<uuid> — viewer identity, for per-client visibility
//	                 filtering. Optional: omitting it yields a
//	                 spectator view where every opponent hand is
//	                 hidden.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Fast path: refuse new connections once shutdown has begun.
	h.mu.RLock()
	closed := h.closed
	h.mu.RUnlock()
	if closed {
		http.Error(w, "server shutting down", http.StatusServiceUnavailable)
		return
	}

	binding, err := h.resolveBinding(r)
	if err != nil {
		// resolveBinding logs nothing — we log here so that auth
		// failures are visible at exactly one level.
		h.log.Warn("ws upgrade rejected", "err", err, "remote", r.RemoteAddr)
		http.Error(w, err.Error(), statusFor(err))
		return
	}
	gameID, playerID := binding.GameID, binding.PlayerID

	// If the client requested a specific game, confirm that game
	// resolves before burning a socket on it. A gameID of uuid.Nil
	// means "no game bound" (ping-only connection) — allowed even
	// without a manager so pre-S04 seams (the healthcheck/ping
	// tests) continue to work unchanged.
	var room *Room
	if gameID != uuid.Nil {
		mgr := h.loadManager()
		if mgr == nil {
			http.Error(w, "server has no room manager", http.StatusServiceUnavailable)
			return
		}
		room = mgr.Get(gameID)
		if room == nil {
			http.Error(w, "game not found", http.StatusNotFound)
			return
		}
		// If a player ID was supplied, validate it is actually a seat
		// in the requested game. A mismatch is almost certainly a
		// client bug; reject the upgrade rather than let the client
		// bind to a non-existent viewer and silently receive
		// spectator views.
		if playerID != uuid.Nil && room.Game.PlayerByID(playerID) == nil {
			http.Error(w, "player not in game", http.StatusForbidden)
			return
		}
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("ws upgrade failed", "err", err, "remote", r.RemoteAddr)
		return
	}
	client := &Client{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte, 16),
		log:      h.log.With("remote", r.RemoteAddr, "game", gameID.String(), "player", playerID.String()),
		gameID:   gameID,
		playerID: playerID,
		readOnly: binding.ReadOnly,
	}

	// Pre-stage the initial snapshot into the client's send channel
	// BEFORE the client becomes visible to broadcastToRoom. This closes
	// the "initial snapshot arrives after a concurrent action's
	// broadcast" race: since the client is not yet in h.clients, no
	// concurrent Apply can broadcast to this channel, so the initial
	// snapshot is guaranteed to be the first frame the new client
	// sees. The channel is freshly-created and buffered at 16, so the
	// non-blocking send inside sendRaw always succeeds here.
	//
	// The pre-staged seq is remembered so the post-admit re-check
	// below can detect a broadcast that landed in the gap between
	// this Snapshot and admit — such a broadcast misses the client
	// (not yet in h.clients) and would otherwise leave it rendering
	// stale state until the next action.
	// stageSnapshot captures the room's current state, filters it for
	// this viewer, and queues it on the client's send channel. Shared
	// by the pre-admit stage and the post-admit catch-up below so the
	// filter/marshal/log sequence lives in one place.
	stageSnapshot := func(label string) (uint64, bool) {
		view, seq, err := room.Snapshot()
		if err != nil {
			client.log.Error("build "+label+" snapshot failed", "err", err)
			return 0, false
		}
		raw, marshalErr := marshalSnapshotFrame(seq, protocol.FilterViewFor(view, viewerIDForFilter(playerID)))
		if marshalErr != nil {
			client.log.Error("marshal "+label+" snapshot failed", "err", marshalErr)
			return 0, false
		}
		client.sendRaw(raw)
		client.log.Debug("staged "+label+" snapshot", "seq", seq)
		return seq, true
	}
	var preStagedSeq uint64
	preStaged := false
	if room != nil {
		preStagedSeq, preStaged = stageSnapshot("initial")
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

	// Post-admit re-check: now that the client is in h.clients, every
	// future Apply broadcast reaches it — but one that fired between
	// the pre-stage Snapshot and admit did not. If the room's seq has
	// advanced past the pre-staged frame, stage a fresh snapshot so
	// the client catches up immediately. The cheap Seq() probe guards
	// the expensive full capture (view build + crash-dump rewrite) —
	// in the common no-race case this costs one mutex hop. Ordering
	// stays correct: the pre-staged frame is already first in the send
	// channel, this one queues behind it, and any concurrent broadcast
	// of the same seq is an idempotent duplicate (see Room.Snapshot's
	// no-bump notes).
	if preStaged && room.Seq() > preStagedSeq {
		stageSnapshot("post-admit catch-up")
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

// resolveBinding extracts the Binding the incoming upgrade request
// should be bound to. If an authorizer is attached, it delegates
// entirely. Otherwise it parses the query string with permissive
// defaults (missing game resolves to the singleton room if any;
// missing player yields the zero UUID, i.e. spectator). The legacy
// no-auth path always returns ReadOnly = false — the read-only
// distinction only exists when an authorizer can certify a session
// as RoleSpectator.
func (h *Hub) resolveBinding(r *http.Request) (Binding, error) {
	if h.authorize != nil {
		return h.authorize.AuthorizeUpgrade(r)
	}
	q := r.URL.Query()
	var gameID uuid.UUID
	if raw := q.Get("game"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return Binding{}, errBadRequest("invalid game id")
		}
		gameID = id
	} else if mgr := h.loadManager(); mgr != nil {
		if only := mgr.Singleton(); only != nil {
			gameID = only.Game.ID
		}
	}
	// gameID == uuid.Nil is allowed here: it means "no game bound".
	// The ServeWS caller permits such connections but rejects any
	// subsequent action frame on them. Pings still work — that's the
	// S01 seam that legacy tests (TestPingRoundTrip etc.) exercise
	// without ever setting up a game.

	var playerID uuid.UUID
	if raw := q.Get("player"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return Binding{}, errBadRequest("invalid player id")
		}
		playerID = id
	}
	return Binding{GameID: gameID, PlayerID: playerID}, nil
}

// errBadRequest is a sentinel wrapper so statusFor can map auth
// errors back to their HTTP status. External packages (auth/lobby)
// wrap their own error shapes via this mechanism too — any error
// the authorizer returns is passed through statusFor unchanged.
type httpStatusError struct {
	code int
	msg  string
}

func (e *httpStatusError) Error() string { return e.msg }

func errBadRequest(msg string) error {
	return &httpStatusError{code: http.StatusBadRequest, msg: msg}
}

// StatusError wraps an error message with an HTTP status code that
// the hub will surface on a failed upgrade. Authorizers use this to
// distinguish "unauthenticated" (401) from "forbidden" (403) from
// "bad request" (400).
func StatusError(code int, msg string) error {
	return &httpStatusError{code: code, msg: msg}
}

// statusFor returns the HTTP status the hub should use when rejecting
// an upgrade with the given error. Unwrapped errors default to 400.
func statusFor(err error) int {
	var s *httpStatusError
	if errors.As(err, &s) {
		return s.code
	}
	return http.StatusBadRequest
}

// viewerIDForFilter converts a WS client's bound playerID into the
// viewerID string that protocol.FilterViewFor expects. A seated
// player's UUID stringifies as-is; a spectator / admin client's
// playerID is uuid.Nil, which stringifies to the all-zero UUID —
// NOT the empty string FilterViewFor documents as "no seat, show
// me the spectator view." Without this coercion, the zero-UUID
// string falls through the is-knower fast-path, every battlefield
// card gets redacted as "unknown", and the spectator sees empty
// tiles where the names and images should be. Bug #191 fix.
func viewerIDForFilter(playerID uuid.UUID) string {
	if playerID == uuid.Nil {
		return ""
	}
	return playerID.String()
}

// BroadcastState fans a lobby-applied state change out to every
// client connected to gameID. HTTP-side mutations (join, deck upload,
// start) bump seq via Room.ApplyExternal, but only the hub can reach
// the connected clients — without this call a player already sitting
// on the game page renders pre-mutation state until the next WS
// action happens to broadcast.
func (h *Hub) BroadcastState(gameID uuid.UUID, seq uint64, view protocol.GameView) {
	h.broadcastToRoom(gameID, seq, view)
}

// broadcastToRoom sends a per-client filtered snapshot frame to every
// currently connected client whose bound gameID matches. The hub's
// read lock prevents unregister from closing any client's send
// channel mid-broadcast, so the sends are always to open channels. A
// full send buffer disconnects the slow client rather than dropping
// the frame silently — see sendRaw for the same policy.
//
// The filter + marshal happens per client because each viewer sees a
// tailored view. At ≤4 clients per room this is microseconds and
// keeps the filter logic strictly server-side. If we ever scale past
// that, lift the marshalling outside this method and cache the bytes
// by player ID.
func (h *Hub) broadcastToRoom(gameID uuid.UUID, seq uint64, view protocol.GameView) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		if c.gameID != gameID {
			continue
		}
		raw, err := marshalSnapshotFrame(seq, protocol.FilterViewFor(view, viewerIDForFilter(c.playerID)))
		if err != nil {
			c.log.Error("marshal broadcast snapshot failed", "err", err)
			continue
		}
		c.sendRaw(raw)
	}
}

// broadcastChat sends a pre-marshalled chat frame to every client
// bound to the given game (including the originating sender, so their
// own client surfaces the canonical server-stamped message rather than
// echoing the unstamped local copy). Mirrors broadcastToRoom but skips
// the per-viewer filter step — chat is public within a room.
func (h *Hub) broadcastChat(gameID uuid.UUID, frame []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		if c.gameID != gameID {
			continue
		}
		c.sendRaw(frame)
	}
}

// marshalSnapshotFrame builds a ready-to-send JSON frame carrying the
// given seq + view as a SnapshotPayload. Extracted so that both the
// initial-snapshot-on-connect path and the broadcast-after-action path
// can share a single frame construction.
func marshalSnapshotFrame(seq uint64, view protocol.GameView) ([]byte, error) {
	payload, err := json.Marshal(protocol.SnapshotPayload{Seq: seq, Game: view})
	if err != nil {
		return nil, err
	}
	return json.Marshal(protocol.Frame{
		V:       protocol.Version,
		Kind:    protocol.KindSnapshot,
		ID:      "",
		Payload: payload,
	})
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

	// gameID is the room this client is bound to for the lifetime of
	// the connection. Set at upgrade time from query string or
	// authorizer output; never changes.
	gameID uuid.UUID

	// playerID is the seat identity used for per-viewer visibility
	// filtering. uuid.Nil means "spectator" — every opponent hand is
	// hidden. Never changes across the connection's lifetime; to
	// switch seats, close and reconnect.
	playerID uuid.UUID

	// readOnly gates incoming `action` frames at the hub. When true,
	// any action / undo dispatch is rejected with bad_request before
	// reaching the room layer. Set from Binding.ReadOnly at upgrade
	// time. RoleSpectator sessions get readOnly=true; admin sessions
	// (including admin-spectator with no ?player=) keep readOnly=false
	// because admins need to drive state on a player's behalf. Added
	// in S11.
	readOnly bool
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
	case protocol.KindChat:
		c.handleChat(frame)
	default:
		c.sendError(frame.ID, protocol.CodeBadRequest,
			"unknown or unsupported kind")
	}
}

// handleChat decodes a client chat frame, re-stamps every server-
// authoritative field (author ID, author name, timestamp), and
// broadcasts the result to every other client bound to the same game.
// Chat does not mutate game state, so it bypasses Room.Apply entirely:
// no seq bump, no snapshot rebuild, no crash-recovery dump entry.
//
// Spectator/admin connections (no playerID) chat under the synthetic
// name "spectator" so the UI can still surface their voice without
// mis-attributing it to a seat.
func (c *Client) handleChat(frame protocol.Frame) {
	if c.gameID == uuid.Nil {
		c.sendError(frame.ID, protocol.CodeBadRequest, "chat requires a game binding")
		return
	}
	if len(frame.Payload) == 0 {
		c.sendError(frame.ID, protocol.CodeBadRequest, "chat frame missing payload")
		return
	}
	var in protocol.ChatPayload
	if err := json.Unmarshal(frame.Payload, &in); err != nil {
		c.sendError(frame.ID, protocol.CodeBadJSON, "chat payload is not valid JSON")
		return
	}
	text := strings.TrimSpace(in.Text)
	if text == "" {
		c.sendError(frame.ID, protocol.CodeBadRequest, "chat text is empty")
		return
	}
	if len(text) > protocol.MaxChatTextLen {
		c.sendError(frame.ID, protocol.CodeBadRequest, "chat text exceeds maximum length")
		return
	}

	// Resolve author identity from server state, ignoring whatever the
	// client sent. A spectator (playerID == uuid.Nil) chats under a
	// synthetic name and an empty AuthorID — the client uses presence
	// of AuthorID to decide whether to apply seat-color styling.
	authorName := "spectator"
	authorID := ""
	if c.playerID != uuid.Nil {
		room := c.hub.resolveRoom(c)
		if room == nil {
			c.sendError(frame.ID, protocol.CodeInternal, "game is no longer available")
			return
		}
		if p := room.Game.PlayerByID(c.playerID); p != nil {
			authorName = p.Name
			authorID = c.playerID.String()
		}
	}

	out := protocol.ChatPayload{
		AuthorID:   authorID,
		AuthorName: authorName,
		Text:       text,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
	payload, err := json.Marshal(out)
	if err != nil {
		c.log.Error("ws marshal chat payload", "err", err)
		c.sendError(frame.ID, protocol.CodeInternal, "failed to encode chat payload")
		return
	}
	stamped, err := json.Marshal(protocol.Frame{
		V:       protocol.Version,
		Kind:    protocol.KindChat,
		ID:      frame.ID,
		Payload: payload,
	})
	if err != nil {
		c.log.Error("ws marshal chat frame", "err", err)
		c.sendError(frame.ID, protocol.CodeInternal, "failed to encode chat frame")
		return
	}
	c.hub.broadcastChat(c.gameID, stamped)
}

// handleAction decodes an action payload, dispatches it to the game
// via the actions package, and broadcasts a fresh snapshot to every
// client bound to the same game on success. Validation or dispatch
// errors come back as error frames to the originating client only.
//
// Dispatch + seq alloc + state capture are serialised under the
// room's mutex via Room.Apply, so two concurrent action frames cannot
// interleave their state changes or broadcast non-monotonic snapshots.
func (c *Client) handleAction(frame protocol.Frame) {
	// S11 spectator gate: read-only connections can't mutate state.
	// Reject before resolving the room so the path is cheap and the
	// client gets a single consistent error code regardless of what
	// they tried to send.
	if c.readOnly {
		c.sendError(frame.ID, protocol.CodeBadRequest, "spectator connections are read-only")
		return
	}

	room := c.hub.resolveRoom(c)
	if room == nil {
		// Either the hub has no manager, or the client's game has
		// been evicted since upgrade. Either way is a server-state
		// problem, not a client mistake.
		c.sendError(frame.ID, protocol.CodeInternal, "game is no longer available")
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

	// `undo` is a room-level operation, not a game mutation: it pops
	// the room's undo stack and restores the previous game state.
	// Special-cased here (rather than routed through actions.Dispatch)
	// because Dispatch only has access to *game.Game, not the room
	// that owns the history. The room layer enforces caller gates
	// (top-of-stack must be your own action; you must have undo
	// budget remaining); admin sessions (playerID == uuid.Nil)
	// bypass both.
	if payload.Type == "undo" {
		view, seq, err := room.Undo(c.playerID)
		if err != nil {
			switch {
			case errors.Is(err, ErrNothingToUndo):
				c.sendError(frame.ID, protocol.CodeBadRequest, "nothing to undo")
			case errors.Is(err, ErrNotYourUndo):
				c.sendError(frame.ID, protocol.CodeBadRequest,
					"you can only undo your own most recent action")
			case errors.Is(err, game.ErrNoUndosRemaining):
				c.sendError(frame.ID, protocol.CodeBadRequest,
					"no undos remaining this turn")
			default:
				c.sendError(frame.ID, protocol.CodeInternal, err.Error())
			}
			return
		}
		c.hub.broadcastToRoom(room.Game.ID, seq, view)
		c.log.Debug("undo applied", "seq", seq, "caller", c.playerID)
		return
	}

	action, err := actions.Decode(payload.Type, payload.Player, payload.Params)
	if err != nil {
		c.sendError(frame.ID, protocol.CodeBadRequest, err.Error())
		return
	}
	// Stamp the authenticated caller so per-seat gates in Dispatch
	// (pass_priority, pass_turn) can validate that the sender is
	// actually the seat allowed to act. uuid.Nil here means admin or
	// spectator — gated branches let those through unchanged.
	action.Caller = c.playerID

	view, seq, err := room.Apply(c.playerID, func() error {
		return actions.Dispatch(room.Game, action)
	})
	if err != nil {
		// S15: the structured insufficient_mana error carries the
		// missing-symbols slice so the client's "Override strict
		// mode for this cast" toast knows what's short. Surface it
		// before the generic classifier flattens the wire.
		var im *game.InsufficientManaError
		if errors.As(err, &im) {
			cardID := ""
			if payload.Type == "cast_spell" {
				var cp struct {
					InstanceID string `json:"instance_id"`
				}
				_ = json.Unmarshal(payload.Params, &cp)
				cardID = cp.InstanceID
			}
			c.sendErrorPayload(frame.ID, protocol.ErrorPayload{
				Code:    protocol.CodeInsufficientMana,
				Message: "insufficient mana to cast",
				Missing: im.Missing,
				CardID:  cardID,
			})
			return
		}
		code, msg := classifyActionError(err)
		c.sendError(frame.ID, code, msg)
		return
	}
	c.hub.broadcastToRoom(room.Game.ID, seq, view)
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
	case errors.Is(err, game.ErrPlayerEliminated):
		return protocol.CodeBadRequest, "player is already eliminated"
	case errors.Is(err, actions.ErrNotPriorityHolder):
		return protocol.CodeBadRequest, "you do not hold priority"
	case errors.Is(err, actions.ErrNotActivePlayer):
		return protocol.CodeBadRequest, "you are not the active player"
	case errors.Is(err, actions.ErrPlayerCallerMismatch):
		return protocol.CodeBadRequest, "you cannot act on another player's behalf"
	case errors.Is(err, game.ErrCardCallerMismatch):
		return protocol.CodeBadRequest, "you do not control that card"
	case errors.Is(err, game.ErrWrongStep):
		return protocol.CodeBadRequest, "action is not legal in the current step"
	case errors.Is(err, game.ErrNotACreature):
		return protocol.CodeBadRequest, "card is not a creature"
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
	c.sendErrorPayload(id, protocol.ErrorPayload{Code: code, Message: message})
}

// sendErrorPayload sends a fully-populated ErrorPayload — used by
// the structured-error paths that need to carry Missing / CardID
// alongside Code + Message (S15 sub-PR 3 insufficient_mana flow).
// Plain code+message callers should keep using sendError for
// brevity; this helper is the escape hatch.
func (c *Client) sendErrorPayload(id string, body protocol.ErrorPayload) {
	payload, err := json.Marshal(body)
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

// EvictGame closes every connected client bound to gameID, sending a
// WebSocket close frame so the peer knows the game is gone (not a
// transport hiccup) and can render a "game ended" state rather than
// reconnect. Returns the number of clients evicted.
//
// Invoked by the lobby when a game is deleted. The read pump's
// deferred unregister handles the map cleanup when each conn closes,
// so we don't touch h.clients under the write lock here — we just
// close sockets and let the existing lifecycle drain.
func (h *Hub) EvictGame(gameID uuid.UUID) int {
	h.mu.RLock()
	victims := make([]*Client, 0)
	for c := range h.clients {
		if c.gameID == gameID {
			victims = append(victims, c)
		}
	}
	h.mu.RUnlock()

	for _, c := range victims {
		_ = c.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "game deleted"),
			time.Now().Add(writeWait),
		)
		_ = c.conn.Close()
	}
	return len(victims)
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
