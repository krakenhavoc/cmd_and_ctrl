package lobby

import (
	"cmp"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// GET /admin/live, the admin views' Live now (ADR 0124 §3.4): who is
// connected now, to which table, as a seat, a spectator or an admin,
// plus the bot and agent seats of every running table.
//
// # Locking
//
// The handler reads the hub's sockets (Config.LiveSockets), then the
// lobby's tables (Lobby.LiveTables), then the database for names. Each
// copy releases its lock before the next starts, in the order the
// metrics tables collector reads them, so the handler never holds two
// of the server's locks at once and takes no room or game lock. The
// two copies are not one instant: a socket that connects between them
// can be counted against a table read after it, the skew ADR 0123
// already accepts for a gauge. The next poll is right.
//
// # Totals
//
// players_connected, spectators, bot_seats and practice_tables are
// metrics.Tally over the same two copies, stripped to the metrics
// types: the function the Overview's tiles are emitted from, so the
// page and Grafana differ only by the time between a scrape and a
// load. admin_views is the number of admin entries the page lists.

// LiveSocketSource is the hub's copy of every live connection, with
// the session's user and connection time (ADR 0124 §5). *ws.Hub
// implements it.
type LiveSocketSource interface {
	LiveSockets() []ws.LiveSocket
}

// Metrics is the table as MetricsTables would have copied it: no
// names, no user IDs, a seat signed in when it has a user.
func (t LiveTable) Metrics() metrics.Table {
	out := metrics.Table{
		Game:     metrics.Key(t.ID),
		State:    t.State,
		Archived: t.Archived,
		Practice: t.Practice,
		Seats:    make([]metrics.Seat, 0, len(t.Seats)),
	}
	for _, s := range t.Seats {
		out.Seats = append(out.Seats, metrics.Seat{Player: metrics.Key(s.PlayerID), Kind: s.Kind, SignedIn: s.UserID != ""})
	}
	return out
}

// liveNowResponse is GET /admin/live's body. The fields here are the
// whole response (ADR 0124 §3); TestAdminLiveServesOnlyItsFields pins
// them.
type liveNowResponse struct {
	GeneratedAt int64          `json:"generated_at"`
	Tables      []liveNowTable `json:"tables"`
	// UnboundSockets counts the sockets whose game the lobby does not
	// hold. It should be 0.
	UnboundSockets int           `json:"unbound_sockets"`
	Totals         liveNowTotals `json:"totals"`
}

type liveNowTable struct {
	ID         uuid.UUID          `json:"id"`
	Name       string             `json:"name"`
	State      string             `json:"state"`
	Practice   bool               `json:"practice"`
	Archived   bool               `json:"archived"`
	Seats      []liveNowSeat      `json:"seats"`
	Spectators []liveNowSpectator `json:"spectators"`
	Admins     []liveNowAdmin     `json:"admins"`
}

// liveNowAccount is a users row as the views show it: never the
// snowflake on its own, only inside the avatar path.
type liveNowAccount struct {
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// liveNowSeat is a seat as in ADR 0124 §3.3, from memory, plus since.
type liveNowSeat struct {
	Seat           int             `json:"seat"`
	Kind           string          `json:"kind"`
	Account        *liveNowAccount `json:"account,omitempty"`
	GuestName      string          `json:"guest_name,omitempty"`
	DiscordPending bool            `json:"discord_pending,omitempty"`
	BotTier        string          `json:"bot_tier,omitempty"`
	AgentClient    string          `json:"agent_client,omitempty"`
	DeckName       string          `json:"deck_name,omitempty"`
	Host           bool            `json:"host"`
	// Connected is the number of live sockets bound to the seat, an
	// admin's included; Since is the earliest of their connection
	// times, absent when there is none.
	Connected int   `json:"connected"`
	Since     int64 `json:"since,omitempty"`
}

// liveNowSpectator is one read-only socket. Account is absent for a
// guest spectator; a spectator socket carries no name of its own.
type liveNowSpectator struct {
	Account *liveNowAccount `json:"account,omitempty"`
	Since   int64           `json:"since"`
}

// liveNowAdmin is one admin socket. Account is absent for the shared
// token; AsSeat is the seat the admin is bound to, if any.
type liveNowAdmin struct {
	Account *liveNowAccount `json:"account,omitempty"`
	AsSeat  *int            `json:"as_seat,omitempty"`
	Since   int64           `json:"since"`
}

type liveNowTotals struct {
	PlayersConnected int `json:"players_connected"`
	Spectators       int `json:"spectators"`
	BotSeats         int `json:"bot_seats"`
	PracticeTables   int `json:"practice_tables"`
	AdminViews       int `json:"admin_views"`
}

// adminLive is GET /admin/live, behind requireAdmin.
func adminLive(c Config, w http.ResponseWriter, r *http.Request) error {
	if c.LiveSockets == nil {
		return httpError(http.StatusServiceUnavailable, "live view needs the WebSocket hub, and this server has none")
	}
	// The hub, then the lobby, then the database: one lock at a time.
	sockets := c.LiveSockets.LiveSockets()
	tables := c.Lobby.LiveTables()

	out, ids := buildLiveNow(tables, sockets)
	names := map[uuid.UUID]users.User{}
	if len(ids) > 0 {
		got, err := c.userStore().Names(r.Context(), ids)
		if err != nil {
			// The page still serves, with the names memory holds.
			c.logger().Error("admin live: reading account names failed", "err", err)
		} else {
			names = got
		}
	}
	out.fillNames(names)
	out.GeneratedAt = c.now().UnixMilli()
	return writeJSON(w, http.StatusOK, out)
}

// buildLiveNow merges the two copies into the response, with accounts
// carrying only their IDs and the seats' own names, and returns every
// account ID it names, for one lookup. It is pure.
func buildLiveNow(tables []LiveTable, sockets []ws.LiveSocket) (liveNowResponse, []uuid.UUID) {
	mt := make([]metrics.Table, 0, len(tables))
	for _, t := range tables {
		mt = append(mt, t.Metrics())
	}
	ms := make([]metrics.Socket, 0, len(sockets))
	for _, s := range sockets {
		ms = append(ms, s.Metrics())
	}
	n := metrics.Tally(mt, ms)
	out := liveNowResponse{
		Tables: []liveNowTable{},
		Totals: liveNowTotals{
			PlayersConnected: n.PlayersConnected(),
			Spectators:       n.Roles[metrics.RoleSpectator],
			BotSeats:         n.Seats[metrics.SeatBot],
			PracticeTables:   n.Practice,
		},
	}

	byGame := map[uuid.UUID][]ws.LiveSocket{}
	held := map[uuid.UUID]bool{}
	for _, t := range tables {
		held[t.ID] = true
	}
	for _, s := range sockets {
		if !held[s.GameID] {
			out.UnboundSockets++
			continue
		}
		byGame[s.GameID] = append(byGame[s.GameID], s)
	}

	var ids []uuid.UUID
	seen := map[uuid.UUID]bool{}
	account := func(id uuid.UUID) *liveNowAccount {
		if id == uuid.Nil {
			return nil
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
		return &liveNowAccount{ID: id.String()}
	}
	since := func(at time.Time) int64 {
		if at.IsZero() {
			return 0
		}
		return at.UnixMilli()
	}

	for _, t := range tables {
		live := byGame[t.ID]
		running := t.State == string(game.StateActive) && !t.Archived
		if len(live) == 0 && !running {
			continue
		}
		row := liveNowTable{
			ID: t.ID, Name: t.Name, State: t.State, Practice: t.Practice, Archived: t.Archived,
			Seats: []liveNowSeat{}, Spectators: []liveNowSpectator{}, Admins: []liveNowAdmin{},
		}
		seatOf := map[uuid.UUID]int{}
		for _, s := range t.Seats {
			seat := liveNowSeat{
				Seat: s.Seat, Kind: s.Kind, DiscordPending: s.DiscordPending,
				BotTier: s.BotTier, AgentClient: s.AgentClient, DeckName: s.DeckName, Host: s.Host,
			}
			if id, err := uuid.Parse(s.UserID); err == nil && id != uuid.Nil {
				seat.Account = account(id)
				seat.Account.Name = cmp.Or(s.DisplayName, s.Name)
			} else {
				seat.GuestName = s.Name
			}
			seatOf[s.PlayerID] = len(row.Seats)
			row.Seats = append(row.Seats, seat)
		}
		for _, s := range live {
			i, seated := seatOf[s.PlayerID]
			if s.PlayerID != uuid.Nil && seated {
				seat := &row.Seats[i]
				seat.Connected++
				if at := since(s.ConnectedAt); at != 0 && (seat.Since == 0 || at < seat.Since) {
					seat.Since = at
				}
			}
			switch {
			case s.Admin || s.Role() == metrics.RoleAdmin:
				a := liveNowAdmin{Account: account(s.UserID), Since: since(s.ConnectedAt)}
				if seated {
					n := row.Seats[i].Seat
					a.AsSeat = &n
				}
				row.Admins = append(row.Admins, a)
			case s.ReadOnly:
				row.Spectators = append(row.Spectators, liveNowSpectator{Account: account(s.UserID), Since: since(s.ConnectedAt)})
			}
		}
		slices.SortStableFunc(row.Spectators, func(a, b liveNowSpectator) int { return cmp.Compare(a.Since, b.Since) })
		slices.SortStableFunc(row.Admins, func(a, b liveNowAdmin) int { return cmp.Compare(a.Since, b.Since) })
		out.Totals.AdminViews += len(row.Admins)
		out.Tables = append(out.Tables, row)
	}

	// Running tables first, then waiting, then ended; by name within.
	rank := map[string]int{string(game.StateActive): 0, string(game.StateLobby): 1}
	order := func(s string) int {
		if r, ok := rank[s]; ok {
			return r
		}
		return 2
	}
	slices.SortFunc(out.Tables, func(a, b liveNowTable) int {
		return cmp.Or(
			cmp.Compare(order(a.State), order(b.State)),
			cmp.Compare(a.Name, b.Name),
			cmp.Compare(a.ID.String(), b.ID.String()),
		)
	})
	return out, ids
}

// fillNames puts the database's avatars on every account the response
// names, and its names on the accounts that have none from memory: a
// seat keeps its seat name (ADR 0124 §5), while a spectator's or an
// admin's socket carries no name, so the database's is the only one.
// An account the database does not know keeps what it has.
func (out *liveNowResponse) fillNames(names map[uuid.UUID]users.User) {
	fill := func(a *liveNowAccount) {
		if a == nil {
			return
		}
		id, err := uuid.Parse(a.ID)
		if err != nil {
			return
		}
		u, ok := names[id]
		if !ok {
			return
		}
		a.Name = cmp.Or(a.Name, u.DisplayName)
		a.AvatarURL = u.AvatarURL
	}
	for ti := range out.Tables {
		t := &out.Tables[ti]
		for i := range t.Seats {
			fill(t.Seats[i].Account)
		}
		for i := range t.Spectators {
			fill(t.Spectators[i].Account)
		}
		for i := range t.Admins {
			fill(t.Admins[i].Account)
		}
	}
}
