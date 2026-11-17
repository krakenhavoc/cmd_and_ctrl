package ws

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/clientip"
)

// A feed is something other than a game that a socket watches: the
// draft (ADR 0222 §5, #5438). A feed socket is a hub Client like any
// other, so it shares the origin check, the connection cap, the pumps,
// the send buffer (for its replies: the feed's own frames are
// coalesced, sendFeed), ping and pong, revocation (EvictUserSessions)
// and Shutdown. It is bound to no game: every game frame kind is refused on
// it, and no game broadcast reaches it.

// Feed is one end of a feed, already bound to whoever the socket is
// (the draft's seat).
type Feed interface {
	// Attach starts the feed. It must send the first frame through out
	// before it returns, so that frame is the socket's first. detach is
	// called once, when the socket goes.
	Attach(out FeedSink) (detach func(), err error)
	// Handle is given every client frame the hub does not answer itself
	// (anything but a ping). A non-nil answer is sent as an error frame
	// echoing the frame's id.
	Handle(frame protocol.Frame) *protocol.ErrorPayload
}

// FeedSink is the socket as a feed sees it. Both methods are safe to
// call from any goroutine and never block.
type FeedSink interface {
	// SendFrame sends the feed's state as one server frame of the given
	// kind. A feed's frames are its state, so they are not queued: one
	// not yet written when the next is sent is replaced by it, and a
	// slow client sees fewer frames instead of being disconnected
	// (#5698).
	SendFrame(kind protocol.Kind, payload any)
	// Close ends the socket with a normal closure and the given reason;
	// the client does not redial.
	Close(reason string)
}

// FeedBinding is who a feed socket belongs to.
type FeedBinding struct {
	// UserID and IssuedAt are the session's, so a revoked sign-in loses
	// its feed sockets with its game sockets.
	UserID   uuid.UUID
	IssuedAt time.Time
	// Seated is true for a socket that already holds a place (a draft
	// seat). Like a seated player on the game socket, it is admitted at
	// the connection cap: its draft is already counted.
	Seated bool
	// Label names the feed in the socket's log lines ("draft <id>").
	Label string
}

// ServeFeed upgrades the request and binds the socket to feed. The
// caller has already authorized it; the hub applies its own gates:
// shutdown, the origin check and the connection cap.
func (h *Hub) ServeFeed(w http.ResponseWriter, r *http.Request, feed Feed, b FeedBinding) {
	h.mu.RLock()
	closed := h.closed
	h.mu.RUnlock()
	if closed {
		metrics.WSUpgradeRejected(metrics.RejectShuttingDown)
		http.Error(w, "server shutting down", http.StatusServiceUnavailable)
		return
	}
	if !b.Seated && h.connectionsFull() {
		h.refuseFull(w, r)
		return
	}
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		metrics.WSUpgradeRejected(metrics.RejectUpgradeFailed)
		h.log.Error("ws feed upgrade failed", "err", err, "remote", clientip.FromRequest(r))
		return
	}
	client := &Client{
		hub:       h,
		conn:      conn,
		send:      make(chan []byte, 16),
		log:       h.log.With("remote", clientip.FromRequest(r), "feed", b.Label),
		userID:    b.UserID,
		issuedAt:  b.IssuedAt,
		feed:      feed,
		feedReady: make(chan struct{}, 1),
	}
	// Attach stages the first frame for the writer (sendFeed) before
	// the client is visible to anything else, as ServeWS pre-stages a
	// snapshot.
	detach, err := feed.Attach(feedSink{client})
	if err != nil {
		metrics.WSUpgradeRejected(metrics.RejectGameNotFound)
		_ = conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, err.Error()),
			time.Now().Add(writeWait),
		)
		_ = conn.Close()
		return
	}
	client.feedDetach = detach
	if !h.admit(client) {
		detach()
		metrics.WSUpgradeRejected(metrics.RejectShuttingDown)
		_ = conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutting down"),
			time.Now().Add(writeWait),
		)
		_ = conn.Close()
		return
	}
	metrics.WSConnected(client.role())
	client.log.Info("ws feed client connected", "total", h.Count())

	go func() {
		defer h.wg.Done()
		client.writePump()
	}()
	go func() {
		defer h.wg.Done()
		client.readPump()
	}()
}

// feedSink is a Client as its feed sees it.
type feedSink struct{ c *Client }

func (s feedSink) SendFrame(kind protocol.Kind, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		s.c.log.Error("ws feed marshal payload", "err", err, "kind", string(kind))
		return
	}
	raw, err := json.Marshal(protocol.Frame{V: protocol.Version, Kind: kind, ID: uuid.NewString(), Payload: body})
	if err != nil {
		s.c.log.Error("ws feed marshal frame", "err", err, "kind", string(kind))
		return
	}
	s.c.sendFeed(kind, raw)
}

func (s feedSink) Close(reason string) {
	_ = s.c.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason),
		time.Now().Add(writeWait),
	)
	_ = s.c.conn.Close()
}

// feedFrame is a feed frame waiting for the writer.
type feedFrame struct {
	kind protocol.Kind
	raw  []byte
}

// sendFeed stages a feed frame for the writer, replacing one it has not
// taken yet (#5698). It does not go through send, so a burst of feed
// frames never fills the send buffer. A feed sends from its own
// goroutines, which can race the socket's teardown; once the socket is
// gone (feedGone, set by unregister under the hub's lock) the frame is
// dropped. Before the client is admitted (the first frame, staged by
// Attach) the writer has not started, and the frame waits for it.
func (c *Client) sendFeed(kind protocol.Kind, raw []byte) {
	c.hub.mu.RLock()
	gone := c.feedGone
	c.hub.mu.RUnlock()
	if gone {
		return
	}
	c.feedMu.Lock()
	c.feedNext = &feedFrame{kind: kind, raw: raw}
	c.feedMu.Unlock()
	select {
	case c.feedReady <- struct{}{}:
	default:
		// The writer has been told already, and takes the newest frame.
	}
}

// writeFeed writes the newest staged feed frame, if there is one. The
// write pump alone calls it.
func (c *Client) writeFeed() error {
	if c.feedReady == nil {
		return nil
	}
	c.feedMu.Lock()
	f := c.feedNext
	c.feedNext = nil
	c.feedMu.Unlock()
	if f == nil {
		return nil
	}
	_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := c.conn.WriteMessage(websocket.TextMessage, f.raw); err != nil {
		return err
	}
	metrics.WSFrame(metrics.FrameOut, string(f.kind))
	metrics.WSSentBytes(string(f.kind), len(f.raw))
	return nil
}

// handleFeedFrame routes a client frame on a feed socket.
func (c *Client) handleFeedFrame(frame protocol.Frame) {
	if frame.Kind == protocol.KindPing {
		c.handlePing(frame)
		return
	}
	if refusal := c.feed.Handle(frame); refusal != nil {
		c.sendErrorPayload(frame.ID, *refusal)
	}
}
