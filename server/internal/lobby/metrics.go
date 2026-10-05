package lobby

import (
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// The lobby's half of ADR 0123 §3: the game counters it records as
// tables start and end, and the accessors the metrics collectors read
// at scrape time. Practice tables are never counted in either.
//
// A table's end is counted once, at the first of:
//
//   - syncStateLocked's first sight of the engine's ended state, with
//     the engine's outcome: win or draw, or closed when it recorded
//     none (migration 0007's NULL);
//   - SetArchived on a table the engine still has running: closed,
//     since archiving is how an operator closes a table (/c2-end);
//   - Delete of a table the engine still has running and nobody has
//     archived: closed.
//
// endCounted holds "once" for the life of the entry, so a table
// archived, unarchived and then won counts as closed and nothing more.
// A restored entry starts counted when its row already has an end or
// an archive, which the previous process counted. A restore point the
// boot abandons is not an end: the file is kept for a later binary to
// bring back.

// recordStartLocked counts the entry's start. Callers hold l.mu and
// call it once, when started_at is stamped.
func (l *Lobby) recordStartLocked(entry *gameEntry) {
	if entry.practice {
		return
	}
	metrics.GameStarted(len(entry.meta.Players))
}

// recordEndLocked counts the entry's end with outcome at, once.
// Callers hold l.mu.
func (l *Lobby) recordEndLocked(entry *gameEntry, outcome string, at time.Time) {
	if entry.practice || entry.endCounted {
		return
	}
	entry.endCounted = true
	length := time.Duration(-1)
	if entry.startedAt != nil {
		length = at.Sub(*entry.startedAt)
	}
	metrics.GameEnded(outcome, length)
}

// recordClosedLocked counts a running table closed by an operator
// (archived or deleted) as an end, if the engine still has it running.
// One that has already ended is syncStateLocked's to count, with its
// own outcome. Callers hold l.mu.
func (l *Lobby) recordClosedLocked(entry *gameEntry) {
	if entry.room.Game.CurrentState() != game.StateActive {
		return
	}
	l.recordEndLocked(entry, metrics.OutcomeClosed, time.Now().UTC())
}

// MetricsTables lists every table for the metrics tables collector,
// practice tables included and marked. It copies under l.mu and reads
// only the entry's cached lifecycle state, so it takes no room or game
// lock and writes nothing (List and Get sync the state with the engine
// and may write the row; a scrape does neither). The cache is synced
// at Start, by watchEnd when a game ends, and on every List and Get.
func (l *Lobby) MetricsTables() []metrics.Table {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]metrics.Table, 0, len(l.games))
	for _, e := range l.games {
		t := metrics.Table{
			Game:     metrics.Key(e.meta.ID),
			State:    e.meta.State,
			Archived: e.meta.Archived(),
			Practice: e.practice,
			Seats:    make([]metrics.Seat, 0, len(e.meta.Players)),
		}
		for _, s := range e.meta.Players {
			t.Seats = append(t.Seats, metrics.Seat{
				Player:   metrics.Key(s.PlayerID),
				Kind:     seatKind(s),
				SignedIn: s.UserID != "",
			})
		}
		out = append(out, t)
	}
	return out
}

// MetricsLiveUsers lists the accounts seated at a table that is active
// now, practice and archived tables excluded, for cmdctrl_users_played.
// An account seated at two tables is listed twice; the collector
// counts distinct ones.
func (l *Lobby) MetricsLiveUsers() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []string
	for _, e := range l.games {
		if e.practice || e.meta.Archived() || e.meta.State != string(game.StateActive) {
			continue
		}
		for _, s := range e.meta.Players {
			if s.UserID != "" {
				out = append(out, s.UserID)
			}
		}
	}
	return out
}

// seatKind is the seat's kind label: a bot, an AI agent's guest seat
// (ADR 0122), or a person.
func seatKind(s SeatInfo) string {
	switch {
	case s.IsBot:
		return metrics.SeatBot
	case s.IsAgent:
		return metrics.SeatAgent
	default:
		return metrics.SeatHuman
	}
}
