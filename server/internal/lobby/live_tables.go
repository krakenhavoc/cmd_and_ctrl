package lobby

import (
	"github.com/google/uuid"
)

// LiveTable is one table the lobby holds in memory, as the admin views
// read it (ADR 0124 §5): the overlay a database row cannot give, and
// the only source for a practice table, which has no row.
//
// Unlike metrics.Table it carries names and user IDs. It is served
// only to an admin, through the admin views, and never reaches the
// metrics package (ADR 0123's rule).
type LiveTable struct {
	ID   uuid.UUID
	Name string
	// State is the lobby's cached lifecycle state, as MetricsTables
	// reads it: "lobby", "active" or "ended".
	State    string
	Archived bool
	Practice bool
	Seats    []LiveSeat
}

// LiveSeat is one seat of a LiveTable.
type LiveSeat struct {
	Seat     int
	PlayerID uuid.UUID
	// Kind is metrics.SeatHuman, SeatBot or SeatAgent: the same three
	// kinds as cmdctrl_seats{kind}.
	Kind string
	// UserID is the signed-in person holding the seat, or "" for a
	// guest, a bot, an agent, or a Discord seat still waiting for its
	// person to sign in again.
	UserID string
	// Name is the seat label (SeatInfo.Name), and DisplayName the
	// Discord name a Discord seat carries (SeatInfo.DisplayName), ""
	// for any other seat.
	Name        string
	DisplayName string
	BotTier     string // a bot's tier, "" otherwise
	AgentClient string // an agent's normalised client name, "" otherwise
	Host        bool
}

// LiveTables copies every table the lobby holds, practice tables
// included and marked, for the admin views (ADR 0124 §5). It is
// MetricsTables' sibling and keeps its rule: it copies under l.mu
// only, reads the entry's cached lifecycle state, takes no room or
// game lock, and writes nothing. A caller that also wants the hub's
// sockets reads them before or after, never while holding this.
func (l *Lobby) LiveTables() []LiveTable {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]LiveTable, 0, len(l.games))
	for _, e := range l.games {
		t := LiveTable{
			ID:       e.meta.ID,
			Name:     e.meta.Name,
			State:    e.meta.State,
			Archived: e.meta.Archived(),
			Practice: e.practice,
			Seats:    make([]LiveSeat, 0, len(e.meta.Players)),
		}
		for _, s := range e.meta.Players {
			seat := LiveSeat{
				Seat:        s.Seat,
				PlayerID:    s.PlayerID,
				Kind:        seatKind(s),
				UserID:      s.UserID,
				Name:        s.Name,
				DisplayName: s.DisplayName,
				Host:        s.IsHost,
			}
			if s.IsBot {
				seat.BotTier = s.BotTier
			}
			if s.IsAgent {
				seat.AgentClient = s.AgentClient
			}
			t.Seats = append(t.Seats, seat)
		}
		out = append(out, t)
	}
	return out
}
