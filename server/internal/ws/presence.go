package ws

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// presence.go — ADR 0162 §2: whether a seat's player is at the table.
//
// A seat is AWAY when no connection is bound to it, or when every one
// that is has said its tab is hidden. The client says so with a
// `visibility` frame on visibilitychange and once on connect; the hub
// keeps the latest value per connection. A connection that never sends
// one counts as visible, so an older client is never treated as away
// while it is connected.
//
// The hub also remembers when each seat last stopped being present (a
// connection closed, or reported hidden), so the notifier can tell a
// player who stepped away a minute ago from one who has been gone for
// ten, and a player who came back for a moment from one who never did.

// seatKey names one seat at one table.
type seatKey struct {
	game uuid.UUID
	seat uuid.UUID
}

// seatLeftPrune bounds Hub.seatLeft: past this many entries, a stamp
// prunes every entry older than seatLeftMaxAge. A pruned seat falls
// back to the caller's own first sighting, which is later and so only
// ever delays a notification.
const (
	seatLeftPrune  = 4096
	seatLeftMaxAge = time.Hour
)

// handleVisibility records a visibility frame. It is never answered:
// a client sends it and moves on. A malformed one gets an error frame,
// and changes nothing.
func (c *Client) handleVisibility(frame protocol.Frame) {
	var in protocol.VisibilityPayload
	if len(frame.Payload) == 0 || bytes.Equal(bytes.TrimSpace(frame.Payload), []byte("null")) {
		c.sendError(frame.ID, protocol.CodeBadRequest, "visibility needs a payload: {\"visible\": true|false}")
		return
	}
	if err := json.Unmarshal(frame.Payload, &in); err != nil {
		c.sendError(frame.ID, protocol.CodeBadJSON, "visibility payload is not valid JSON")
		return
	}
	if in.Visible == nil {
		c.sendError(frame.ID, protocol.CodeBadRequest, "visibility needs \"visible\": true or false")
		return
	}
	c.hub.setHidden(c, !*in.Visible)
}

// setHidden records c's latest visibility, under h.mu so SeatAway
// reads it and the seat's departure stamp together.
func (h *Hub) setHidden(c *Client, hidden bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if hidden && !c.hidden {
		h.markLeftLocked(c)
	}
	c.hidden = hidden
}

// markLeftLocked stamps the moment c stopped making its seat present.
// Caller holds h.mu for writing. A connection with no seat, or a
// read-only one, never made a seat present.
func (h *Hub) markLeftLocked(c *Client) {
	if c.playerID == uuid.Nil || c.readOnly || c.gameID == uuid.Nil {
		return
	}
	now := h.now()
	if h.seatLeft == nil {
		h.seatLeft = make(map[seatKey]time.Time)
	}
	if len(h.seatLeft) >= seatLeftPrune {
		for k, at := range h.seatLeft {
			if now.Sub(at) > seatLeftMaxAge {
				delete(h.seatLeft, k)
			}
		}
	}
	h.seatLeft[seatKey{c.gameID, c.playerID}] = now
}

// SeatAway reports whether seat at game is away (ADR 0162 §2): no
// connection bound to it, or every one hidden. since is when the seat
// last stopped being present, and zero when this hub has not seen it
// leave (it was never connected since the server started). A present
// seat answers false and a zero since.
//
// It reads under the hub's read lock and takes no other lock.
func (h *Hub) SeatAway(game, seat uuid.UUID) (away bool, since time.Time) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		if c.gameID == game && c.playerID == seat && !c.readOnly && !c.hidden {
			return false, time.Time{}
		}
	}
	return true, h.seatLeft[seatKey{game, seat}]
}
