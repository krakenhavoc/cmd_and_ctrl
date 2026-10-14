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
// reads it and the seat's departure stamp together. It reports whether
// the visibility changed.
func (h *Hub) setHidden(c *Client, hidden bool) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if hidden && !c.hidden {
		h.markLeftLocked(c)
	}
	changed := c.hidden != hidden
	c.hidden = hidden
	return changed
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

// markOfflineLocked stamps the moment one of c's seat's connections
// closed. Caller holds h.mu for writing. Every close restamps it, so
// once the last one has closed it is when the seat went offline (ADR
// 0177 §1).
func (h *Hub) markOfflineLocked(c *Client) {
	if c.playerID == uuid.Nil || c.readOnly || c.gameID == uuid.Nil {
		return
	}
	now := h.now()
	if h.seatOffline == nil {
		h.seatOffline = make(map[seatKey]time.Time)
	}
	pruneStamps(h.seatOffline, now)
	h.seatOffline[seatKey{c.gameID, c.playerID}] = now
}

// pruneStamps bounds one of the hub's per-seat stamp maps the way
// markLeftLocked bounds seatLeft.
func pruneStamps(m map[seatKey]time.Time, now time.Time) {
	if len(m) < seatLeftPrune {
		return
	}
	for k, at := range m {
		if now.Sub(at) > seatLeftMaxAge {
			delete(m, k)
		}
	}
}

// SeatPresence is ADR 0177 §1's presence of one seat: "" while it is
// present, protocol.PresenceOffline while no connection is bound to
// it, protocol.PresenceHidden while every one that is reports its tab
// hidden. since is when it stopped being present: for offline, when
// its last connection closed; for hidden, when it went hidden. A seat
// the hub has never seen leave (it has not been connected since the
// server started) is away since the hub first answered for it, which
// is how the notifier counts it too.
func (h *Hub) SeatPresence(game, seat uuid.UUID) (state string, since time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.seatPresenceLocked(game, seat)
}

// seatPresenceLocked is SeatPresence. Caller holds h.mu for writing,
// because a first sighting is recorded.
func (h *Hub) seatPresenceLocked(game, seat uuid.UUID) (string, time.Time) {
	bound := false
	for c := range h.clients {
		if c.gameID != game || c.playerID != seat || c.readOnly {
			continue
		}
		if !c.hidden {
			return "", time.Time{}
		}
		bound = true
	}
	k := seatKey{game, seat}
	state, since := protocol.PresenceOffline, h.seatOffline[k]
	if bound {
		state, since = protocol.PresenceHidden, h.seatLeft[k]
	} else if since.IsZero() {
		since = h.seatLeft[k]
	}
	if since.IsZero() {
		if h.seatSeen == nil {
			h.seatSeen = make(map[seatKey]time.Time)
		}
		seen, ok := h.seatSeen[k]
		if !ok {
			seen = h.now()
			pruneStamps(h.seatSeen, seen)
			h.seatSeen[k] = seen
		}
		since = seen
	}
	return state, since
}

// stampPresence returns view with every seat's presence on it (ADR
// 0177 §1), and the room's movable mark (§2) on an offline seat. The
// seats are copied, so the caller's view is untouched. A bot seat and
// a seat that has left the game never carry presence. room may be nil
// (a hub with no manager): presence is still stamped, never movable.
//
// It takes the hub's write lock, so a caller must not hold it.
func (h *Hub) stampPresence(room *Room, view protocol.GameView) protocol.GameView {
	return h.stampPresenceExcept(room, view, uuid.Nil, true)
}

// stampPresenceExcept is stampPresence for the frame a connection to
// self is about to receive before it is admitted: self reads as
// present unless the connection is read-only.
func (h *Hub) stampPresenceExcept(room *Room, view protocol.GameView, self uuid.UUID, readOnly bool) protocol.GameView {
	if len(view.Seats) == 0 {
		return view
	}
	gameID, err := uuid.Parse(view.ID)
	if err != nil {
		return view
	}
	seats := make([]protocol.PlayerView, len(view.Seats))
	copy(seats, view.Seats)
	h.mu.Lock()
	for i := range seats {
		seats[i].Presence = nil
		if seats[i].IsBot || seats[i].Eliminated {
			continue
		}
		id, err := uuid.Parse(seats[i].ID)
		if err != nil || (id == self && !readOnly) {
			continue
		}
		state, since := h.seatPresenceLocked(gameID, id)
		if state == "" {
			continue
		}
		seats[i].Presence = &protocol.PresenceView{State: state, Since: since.UTC().Format(time.RFC3339)}
	}
	h.mu.Unlock()
	if room != nil {
		for i := range seats {
			p := seats[i].Presence
			if p == nil || p.State != protocol.PresenceOffline {
				continue
			}
			if id, err := uuid.Parse(seats[i].ID); err == nil && room.Movable(id) {
				p.Movable = true
			}
		}
	}
	view.Seats = seats
	return view
}

// RefreshPresence sends every viewer of gameID a fresh frame at the
// room's current seq, for a change that is not a commit: a seat's
// presence, or what the lobby says about the table's away seats (ADR
// 0177). The seq is not advanced, because the game did not change; a
// client takes a frame at its current seq. A game with no live room
// sends nothing.
//
// The hub never calls it on its own when a connection opens or closes:
// the lobby's away clock does, on its next tick, for a table whose
// presence changed (lobby.TickAway). So a reload costs the table one
// frame at most, and a test reading frames in order sees none.
func (h *Hub) RefreshPresence(gameID uuid.UUID) {
	m := h.loadManager()
	if m == nil {
		return
	}
	room := m.Get(gameID)
	if room == nil {
		return
	}
	view, seq, err := room.Snapshot()
	if err != nil {
		h.log.Warn("presence refresh: snapshot failed", "game", gameID.String(), "err", err)
		return
	}
	h.broadcastToRoom(gameID, seq, room.Generation(), view)
}
