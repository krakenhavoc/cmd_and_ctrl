package ws

// idle.go: the room's half of the idle clocks (ADR 0211 §1). The lobby
// closes a table nobody has touched for a day and a game nobody has
// moved in for two (lobby/idle.go); what counts as a move is decided
// here, where every commit passes.
//
// A started game's clock is reset only by a person's own move: a commit
// for a seat, at that seat's request, that is not an admin's, not a
// bot's and not a stand-in's. noteActed (ADR 0162 §2) cannot be reused
// as is: a stand-in runner commits as the person's seat, and seatKind
// reads that seat as human. So the room keeps one more stamp beside it.

import (
	"time"

	"github.com/google/uuid"
)

// SetClock replaces the idle clocks' time source. Tests only; nil is
// time.Now.
func (r *Room) SetClock(now func() time.Time) {
	r.actMu.Lock()
	defer r.actMu.Unlock()
	r.clock = now
}

// nowLocked is the idle clocks' time. Caller holds actMu.
func (r *Room) nowLocked() time.Time {
	if r.clock != nil {
		return r.clock()
	}
	return time.Now()
}

// humanMove reports whether a commit tagged tag is a person's (or an
// agent's) own move for their seat: not the lobby or the admin
// (uuid.Nil, or an admin connection bound to a seat), not a bot seat,
// and not a seat a stand-in bot is playing. The hub refuses a stand-in
// seat's own actions until Take it back (ADR 0177 §4), so every commit
// for such a seat is the runner's.
func (r *Room) humanMove(tag ActionTag) bool {
	if tag.Actor == uuid.Nil || tag.Admin {
		return false
	}
	if bot, _ := r.Game.SeatDriver(tag.Actor); bot {
		return false
	}
	return !r.StandIn(tag.Actor)
}

// noteHuman stamps LastHumanAt when the commit tagged tag is a human
// move. Called after the commit, outside r.mu.
func (r *Room) noteHuman(tag ActionTag) {
	if !r.humanMove(tag) {
		return
	}
	r.actMu.Lock()
	defer r.actMu.Unlock()
	r.lastHumanAt = r.nowLocked()
}

// LastHumanAt is when a person or an agent last made a move of their
// own at this table in this process, and zero if nobody has.
func (r *Room) LastHumanAt() time.Time {
	r.actMu.Lock()
	defer r.actMu.Unlock()
	return r.lastHumanAt
}

// noteSeatedChat stamps LastSeatedChatAt for a chat line from seat. A
// bot's seat is not stamped. The hub has already ruled out spectators
// and admins.
func (r *Room) noteSeatedChat(seat uuid.UUID) {
	if bot, _ := r.Game.SeatDriver(seat); bot || seat == uuid.Nil {
		return
	}
	r.actMu.Lock()
	defer r.actMu.Unlock()
	r.lastChatAt = r.nowLocked()
}

// LastSeatedChatAt is when a seated person last sent a chat line in
// this process, and zero if nobody has. Only an unstarted table's clock
// reads it: chat is not a move (ADR 0211, "Calls made here").
func (r *Room) LastSeatedChatAt() time.Time {
	r.actMu.Lock()
	defer r.actMu.Unlock()
	return r.lastChatAt
}
