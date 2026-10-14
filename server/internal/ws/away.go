package ws

// away.go: the room's half of ADR 0177, moving the table on without a
// player who is gone.
//
// Two things live here, both set by the lobby and neither part of the
// game: which seats a bot is standing in for (§4), and which seats the
// table may move on without right now (§2). Like the host and the
// rematch offer they are not in the engine, its snapshot or its undo
// ring. The stand-ins are stamped on every captured view; the movable
// mark rides the presence the hub stamps at send time (presence.go).

import (
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

type standIn struct {
	tier  string
	since time.Time
}

// SetStandIn records that a bot of tier is playing seat, since since.
func (r *Room) SetStandIn(seat uuid.UUID, tier string, since time.Time) {
	r.awayMu.Lock()
	defer r.awayMu.Unlock()
	if r.standIns == nil {
		r.standIns = make(map[uuid.UUID]standIn)
	}
	r.standIns[seat] = standIn{tier: tier, since: since}
	delete(r.movable, seat)
}

// ClearStandIn records that seat's player has it back.
func (r *Room) ClearStandIn(seat uuid.UUID) {
	r.awayMu.Lock()
	defer r.awayMu.Unlock()
	delete(r.standIns, seat)
}

// StandIn reports whether a bot is playing seat for its player.
func (r *Room) StandIn(seat uuid.UUID) bool {
	r.awayMu.Lock()
	defer r.awayMu.Unlock()
	_, ok := r.standIns[seat]
	return ok
}

// SetMovable replaces the set of seats the table may move on without,
// and reports whether it changed.
func (r *Room) SetMovable(seats map[uuid.UUID]bool) bool {
	r.awayMu.Lock()
	defer r.awayMu.Unlock()
	next := make(map[uuid.UUID]bool, len(seats))
	for id, ok := range seats {
		if ok {
			next[id] = true
		}
	}
	changed := len(next) != len(r.movable)
	if !changed {
		for id := range next {
			if !r.movable[id] {
				changed = true
				break
			}
		}
	}
	r.movable = next
	return changed
}

// Movable reports whether the table may move on without seat.
func (r *Room) Movable(seat uuid.UUID) bool {
	r.awayMu.Lock()
	defer r.awayMu.Unlock()
	return r.movable[seat]
}

// stampStandInsLocked writes the stand-ins onto a freshly captured
// view. Caller holds r.mu.
func (r *Room) stampStandInsLocked(view *protocol.GameView) {
	r.awayMu.Lock()
	defer r.awayMu.Unlock()
	if len(r.standIns) == 0 {
		return
	}
	for i := range view.Seats {
		id, err := uuid.Parse(view.Seats[i].ID)
		if err != nil {
			continue
		}
		if s, ok := r.standIns[id]; ok {
			view.Seats[i].StandIn = &protocol.StandInView{Tier: s.tier, Since: s.since.UTC().Format(time.RFC3339)}
		}
	}
}
