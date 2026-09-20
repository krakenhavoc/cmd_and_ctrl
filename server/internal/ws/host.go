package ws

// host.go: the table host (ADR 0075 §2.1) as the room sees it.
//
// The lobby decides WHO is designated host — the first human seat, a
// named Discord identity claiming a seat, or an explicit transfer —
// and tells the room through SetHost. The room owns the one thing the
// lobby cannot see without reaching into the engine: whether that seat
// is still in the game. A host who concedes or loses (ADR 0060) passes
// hosting to the next human seat in turn order, and that is evaluated
// here, against the same game state every commit captures, so the
// is_host flag on the wire, the lobby's GameMeta and any WebSocket
// gate all read one answer.
//
// The pass is sticky: once hosting has moved it is recorded as the new
// designated host, so a later undo that restores the old seat does
// not silently hand the table back.

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// hostSeat is the slice of a seat the host rules read: identity, and
// the two properties that disqualify a seat from hosting.
type hostSeat struct {
	id         uuid.UUID
	bot        bool
	eliminated bool
}

func (s hostSeat) eligible() bool { return !s.bot && !s.eliminated }

// pickHost is the pure host rule. seats are in turn order (seat
// index order — the order the turn rotation walks).
//
//   - No designated host: no host. The lobby designates one when the
//     first human sits down; the room never invents one.
//   - The designated seat is present and eligible: it hosts.
//   - Otherwise hosting passes to the next eligible seat in turn
//     order after the designated one, wrapping around. A designated
//     seat that is not at the table at all starts the walk from seat
//     0.
//   - Nobody eligible: no host. The admin can still manage the table.
func pickHost(designated uuid.UUID, seats []hostSeat) uuid.UUID {
	if designated == uuid.Nil || len(seats) == 0 {
		return uuid.Nil
	}
	start := -1
	for i, s := range seats {
		if s.id == designated {
			if s.eligible() {
				return s.id
			}
			start = i
			break
		}
	}
	for step := 1; step <= len(seats); step++ {
		s := seats[(start+step+len(seats))%len(seats)]
		if s.eligible() {
			return s.id
		}
	}
	return uuid.Nil
}

// SetHost records the designated host. uuid.Nil clears it. The lobby
// calls it on the first human join, on a named host's claim, on a
// transfer, and when a restored room is paired with its metadata.
//
// Safe to call from inside an ApplyExternal fn (it takes only the
// host mutex), which is how a join makes its own capture carry the
// new is_host flag.
func (r *Room) SetHost(playerID uuid.UUID) {
	r.hostMu.Lock()
	r.host = playerID
	r.hostMu.Unlock()
}

// HostPlayerID returns the effective host: the designated host if it
// is still an eligible seat, else whoever hosting has passed to.
// uuid.Nil means the table has no host (only the admin can manage
// it). Takes the game's read lock; do not call it while holding the
// game lock.
func (r *Room) HostPlayerID() uuid.UUID {
	var seats []hostSeat
	r.Game.ReadSnapshot(func() {
		seats = make([]hostSeat, 0, len(r.Game.Seats))
		for _, p := range r.Game.Seats {
			if p == nil {
				continue
			}
			seats = append(seats, hostSeat{id: p.ID, bot: p.IsBot, eliminated: p.Eliminated})
		}
	})
	return r.resolveHost(seats)
}

// resolveHost applies pickHost to the designated host and makes the
// result sticky.
func (r *Room) resolveHost(seats []hostSeat) uuid.UUID {
	r.hostMu.Lock()
	defer r.hostMu.Unlock()
	h := pickHost(r.host, seats)
	r.host = h
	return h
}

// IsHost reports whether playerID is the table's effective host.
func (r *Room) IsHost(playerID uuid.UUID) bool {
	return playerID != uuid.Nil && r.HostPlayerID() == playerID
}

// CanManageTable is the WebSocket-side twin of lobby.CanManageTable
// (ADR 0075 §2.1): true for an admin connection, or for a seated,
// non-read-only connection bound to the table's effective host.
// Nothing calls it yet — the action gates arrive with ADR 0075
// sub-PR 3 — but the answer lives here so the HTTP and WebSocket
// sides cannot drift apart.
func (r *Room) CanManageTable(b Binding) bool {
	if b.Admin {
		return true
	}
	if b.ReadOnly || b.PlayerID == uuid.Nil {
		return false
	}
	if b.GameID != uuid.Nil && r.Game != nil && b.GameID != r.Game.ID {
		return false
	}
	return r.IsHost(b.PlayerID)
}

// stampHostLocked marks the effective host on a freshly captured view
// and records any pass. Caller holds r.mu; the view was built from the
// game state the commit just produced, so reading the seats off it
// needs no second game lock.
func (r *Room) stampHostLocked(view *protocol.GameView) {
	seats := make([]hostSeat, 0, len(view.Seats))
	for _, s := range view.Seats {
		id, err := uuid.Parse(s.ID)
		if err != nil {
			continue
		}
		seats = append(seats, hostSeat{id: id, bot: s.IsBot, eliminated: s.Eliminated})
	}
	h := r.resolveHost(seats)
	if h == uuid.Nil {
		return
	}
	hs := h.String()
	for i := range view.Seats {
		view.Seats[i].IsHost = view.Seats[i].ID == hs
	}
	// The public log's spawn lines say whether the spawner held the
	// table (ADR 0075 §2.4). The projection cannot know that — the
	// host is ours, not the engine's — so the same pass that stamps
	// the seats stamps the log.
	protocol.StampHostOnLog(view)
}
