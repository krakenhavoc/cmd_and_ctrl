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

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
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

// noteHuman stamps LastHumanAt when the commit was a human move (the
// commit's humanMove, read before it took r.mu). Called after the
// commit, outside r.mu.
func (r *Room) noteHuman(human bool) {
	if !human {
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

// SetIdleCloseAt is the lobby's warning that this started game closes
// at closesAt unless a person moves (ADR 0211 §6); zero takes it back.
// Every captured view carries it as idle_close_at until the next human
// move clears it, in the same commit, so the frame that answers the
// warning no longer shows it. Like SetRematch it takes only its own
// mutex, so the lobby may call it under l.mu or inside an ApplyExternal
// fn.
func (r *Room) SetIdleCloseAt(closesAt time.Time) {
	r.idleMu.Lock()
	defer r.idleMu.Unlock()
	r.idleCloseAt = closesAt
}

// IdleCloseAt is the warning's closing time, zero when the game is not
// warned.
func (r *Room) IdleCloseAt() time.Time {
	r.idleMu.Lock()
	defer r.idleMu.Unlock()
	return r.idleCloseAt
}

// clearIdleWarning takes the warning back when a commit is a human
// move. Called under r.mu after the commit's fn succeeded and before its
// capture, so the commit's own frame no longer carries it.
func (r *Room) clearIdleWarning(human bool) {
	if human {
		r.SetIdleCloseAt(time.Time{})
	}
}

// stampIdleLocked writes the warning onto a freshly captured view.
// Caller holds r.mu.
func (r *Room) stampIdleLocked(view *protocol.GameView) {
	if at := r.IdleCloseAt(); !at.IsZero() {
		view.IdleCloseAt = at.UTC().Format(time.RFC3339)
	}
}

// SetPlayOut turns on or off the fast play-out of a game left to its
// bots (ADR 0211 §2, owner decision 2): while it is on, every bot and
// stand-in runner at the table plays at game.BotPaceFast whatever the
// table's own pace, so the game reaches a real result rather than
// waiting on the one-hour close. The lobby's sweep turns it on for a
// bot-only game nobody is watching, and off when that stops being so; a
// person's seat connecting turns it off at once (Hub.ServeWS). It lives
// in memory only, outside the engine and its settings.
func (r *Room) SetPlayOut(on bool) { r.playOut.Store(on) }

// PlayingOut reports whether the bots are playing the game out fast.
func (r *Room) PlayingOut() bool { return r.playOut.Load() }
