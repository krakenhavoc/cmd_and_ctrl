package lobby

// notify.go is the lobby's half of ADR 0162's DM notifications: it
// tells the notifier when a table fills or starts, and answers its
// questions about the tables (notify.Tables). The notifier decides who
// is told; the lobby only reports what happened.

import (
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/notify"
)

// TableEvents is told when a table fills or starts. Satisfied by
// *notify.Notifier; nil (no bot token) tells nobody.
type TableEvents interface {
	// TableFilled: every seat at the table is taken. by is the seat
	// whose join filled it, uuid.Nil for a bot.
	TableFilled(game, by uuid.UUID)
	// TableStarted: the table has started.
	TableStarted(game uuid.UUID)
}

// SetTableEvents wires the notifier in at boot.
func (l *Lobby) SetTableEvents(e TableEvents) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = e
}

// noteFilledLocked tells the notifier entry has just filled. Caller
// holds l.mu; the notifier only queues, and never calls back into the
// lobby from there.
func (l *Lobby) noteFilledLocked(entry *gameEntry, by uuid.UUID) {
	if l.events == nil || entry.practice || len(entry.meta.Players) < game.MaxPlayers {
		return
	}
	l.events.TableFilled(entry.meta.ID, by)
}

// noteStartedLocked tells the notifier entry has just started.
func (l *Lobby) noteStartedLocked(entry *gameEntry) {
	if l.events == nil || entry.practice {
		return
	}
	l.events.TableStarted(entry.meta.ID)
}

// NotifyTables is the lobby as notify.Tables.
func (l *Lobby) NotifyTables() notify.Tables { return notifyTables{l} }

type notifyTables struct{ l *Lobby }

// Live implements notify.Tables: the tables in play, not archived and
// not practice.
func (n notifyTables) Live() []notify.Table {
	l := n.l
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []notify.Table
	for _, e := range l.games {
		if e.practice || e.meta.Archived() || e.room == nil || e.room.Game.CurrentState() != game.StateActive {
			continue
		}
		out = append(out, notifyTableLocked(e))
	}
	return out
}

// Table implements notify.Tables.
func (n notifyTables) Table(id uuid.UUID) (notify.Table, bool) {
	l := n.l
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.games[id]
	if !ok || e.practice || e.meta.Archived() || e.room == nil {
		return notify.Table{}, false
	}
	return notifyTableLocked(e), true
}

// Waiting implements notify.Tables: the engine waits on a seat when
// that seat has a legal move, the same enumeration the seat's own
// legal_moves come from (ADR 0105). It holds the priority, owes a
// prompt, a mulligan, a roll, a discard or a combat declaration.
func (n notifyTables) Waiting(gameID, seat uuid.UUID) (bool, time.Time) {
	room := n.l.RoomOf(gameID)
	if room == nil {
		return false, time.Time{}
	}
	return len(legal.EnumerateFor(room.Game, seat)) > 0, room.LastActed(seat)
}

// notifyTableLocked projects one entry. Caller holds l.mu.
func notifyTableLocked(e *gameEntry) notify.Table {
	t := notify.Table{
		ID:    e.meta.ID,
		Name:  e.meta.Name,
		State: string(e.room.Game.CurrentState()),
		Full:  len(e.meta.Players) >= game.MaxPlayers,
	}
	if glance, ok := e.room.Game.Glance(); ok {
		t.Turn = glance.Turn
	}
	host := e.room.HostPlayerID()
	for _, s := range e.meta.Players {
		t.Seats = append(t.Seats, notify.Seat{
			PlayerID: s.PlayerID,
			UserID:   parseUUIDOrNil(s.UserID),
			Host:     s.PlayerID == host && host != uuid.Nil,
			Bot:      s.IsBot,
			Agent:    s.IsAgent,
		})
	}
	return t
}
