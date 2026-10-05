package adminview

import (
	"cmp"
	"sort"
	"time"

	"github.com/google/uuid"
)

// --- the live overlay ------------------------------------------------

// LiveTable is one table the lobby holds in memory, as lobby.LiveTable
// copies it: the overlay a row cannot give, and the only source for a
// practice table, which has no row.
type LiveTable struct {
	ID        uuid.UUID
	Name      string
	State     string
	CreatedAt time.Time
	Archived  bool
	Practice  bool
	Seats     []LiveSeat
}

// LiveSeat is one seat of a LiveTable.
type LiveSeat struct {
	Seat        int
	PlayerID    uuid.UUID
	Kind        string // human | bot | agent
	UserID      string
	Name        string
	DisplayName string
	BotTier     string
	AgentClient string
	DeckName    string
	Host        bool
	// DiscordPending is a Discord seat whose person has no users row
	// yet; the snowflake is not carried.
	DiscordPending bool
}

// Socket is one live WebSocket, as far as the table rows need it.
type Socket struct {
	GameID   uuid.UUID
	PlayerID uuid.UUID // uuid.Nil for a connection with no seat
	// UserID is the signed-in person behind the session, "" for a
	// guest, a guest spectator or the shared admin token.
	UserID      string
	ReadOnly    bool
	Admin       bool
	ConnectedAt time.Time
}

// spectator is GET /admin/live's rule: a read-only socket that is not
// an admin's.
func (s Socket) spectator() bool { return s.ReadOnly && !s.Admin }

// admin is GET /admin/live's rule: an admin-bound socket, or one with
// no seat that is not read-only (metrics.WSRole's admin).
func (s Socket) admin() bool { return s.Admin || (s.PlayerID == uuid.Nil && !s.ReadOnly) }

// Overlay is what memory knows, copied before the database is read.
type Overlay struct {
	Tables []LiveTable
	// Sockets are the hub's live sockets, read only when SocketsKnown.
	// Without them a loaded table's connected counts are absent, never
	// a zero that looks like a fact.
	Sockets      []Socket
	SocketsKnown bool
	// Accounts names the accounts seated at tables that have no row
	// (Store.AccountRefs).
	Accounts map[string]AccountRef
}

func (o Overlay) table(id uuid.UUID) (LiveTable, bool) {
	for _, t := range o.Tables {
		if t.ID == id {
			return t, true
		}
	}
	return LiveTable{}, false
}

// PracticeUserIDs is the accounts seated at the overlay's practice
// tables: the IDs a caller passes to Store.AccountRefs before listing
// them.
func (o Overlay) PracticeUserIDs() []string {
	var out []string
	for _, t := range o.Tables {
		if !t.Practice {
			continue
		}
		for _, s := range t.Seats {
			if s.UserID != "" {
				out = append(out, s.UserID)
			}
		}
	}
	return out
}

// --- responses -------------------------------------------------------

// AccountRef is an account as a seat names it.
type AccountRef struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// Creator is the account that created a table.
type Creator struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Account is one account row (ADR 0124 §3.1).
type Account struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	AvatarURL    string `json:"avatar_url,omitempty"`
	FirstSeenAt  int64  `json:"first_seen_at,omitempty"`
	LastSignInAt int64  `json:"last_sign_in_at,omitempty"`
	GamesPlayed  int    `json:"games_played"`
	LastPlayedAt int64  `json:"last_played_at,omitempty"`
	PlayingNow   bool   `json:"playing_now"`
}

// AccountsResponse is GET /admin/users.
type AccountsResponse struct {
	GeneratedAt int64     `json:"generated_at"`
	Accounts    []Account `json:"accounts"`
	Truncated   bool      `json:"truncated"`
}

// SignIn is an account's sign-in state (§3.2). Sessions are HMAC
// tokens with no row, so there is no list of them.
type SignIn struct {
	LastSignInAt          int64  `json:"last_sign_in_at,omitempty"`
	DiscordLinkedAt       int64  `json:"discord_linked_at,omitempty"`
	SessionsInvalidBefore int64  `json:"sessions_invalid_before,omitempty"`
	RevokePath            string `json:"revoke_path"`
}

// Deck is a saved deck without its list (§3.2).
type Deck struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Format     string   `json:"format"`
	SourceURL  string   `json:"source_url,omitempty"`
	Commanders []string `json:"commanders"`
	CardCount  int      `json:"card_count"`
	CreatedAt  int64    `json:"created_at,omitempty"`
	UpdatedAt  int64    `json:"updated_at,omitempty"`
}

// DeckRequest is one ask (§3.2).
type DeckRequest struct {
	DeckKey     string `json:"deck_key"`
	AskedAt     int64  `json:"asked_at,omitempty"`
	IssueNumber int    `json:"issue_number,omitempty"`
	IssueURL    string `json:"issue_url,omitempty"`
}

// AccountResponse is GET /admin/users/{id}.
type AccountResponse struct {
	GeneratedAt           int64         `json:"generated_at"`
	Account               Account       `json:"account"`
	SignIn                SignIn        `json:"sign_in"`
	Games                 []Game        `json:"games"`
	GamesTruncated        bool          `json:"games_truncated"`
	Decks                 []Deck        `json:"decks"`
	DeckRequests          []DeckRequest `json:"deck_requests"`
	DeckRequestsTruncated bool          `json:"deck_requests_truncated"`
}

// Game is one table row (§3.3).
type Game struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	State      string   `json:"state"`
	CreatedAt  int64    `json:"created_at,omitempty"`
	StartedAt  int64    `json:"started_at,omitempty"`
	EndedAt    int64    `json:"ended_at,omitempty"`
	ArchivedAt int64    `json:"archived_at,omitempty"`
	Outcome    string   `json:"outcome,omitempty"`
	WinnerSeat *int     `json:"winner_seat,omitempty"`
	Creator    *Creator `json:"creator,omitempty"`
	Practice   bool     `json:"practice"`
	Loaded     bool     `json:"loaded"`
	Seats      []Seat   `json:"seats"`
	// SpectatorsConnected is the table's live read-only sockets: loaded
	// tables only, and only when the server knows its sockets.
	SpectatorsConnected *int `json:"spectators_connected,omitempty"`
	// TheirSeat is the account's seat, on an account's games only.
	TheirSeat *int `json:"their_seat,omitempty"`
}

// Seat is one seat of a table row (§3.3).
type Seat struct {
	Seat           int         `json:"seat"`
	Kind           string      `json:"kind"`
	Account        *AccountRef `json:"account,omitempty"`
	GuestName      string      `json:"guest_name,omitempty"`
	DiscordPending bool        `json:"discord_pending,omitempty"`
	BotTier        string      `json:"bot_tier,omitempty"`
	AgentClient    string      `json:"agent_client,omitempty"`
	DeckName       string      `json:"deck_name,omitempty"`
	Host           bool        `json:"host"`
	// Connected is the live sockets bound to this seat: loaded tables
	// only, and only when the server knows its sockets.
	Connected *int `json:"connected,omitempty"`
}

// GamesResponse is GET /admin/games.
type GamesResponse struct {
	GeneratedAt int64  `json:"generated_at"`
	Games       []Game `json:"games"`
	NextCursor  string `json:"next_cursor,omitempty"`
}

// GameResponse is GET /admin/games/{id}: the table's row, flattened,
// and its live connections.
type GameResponse struct {
	GeneratedAt int64 `json:"generated_at"`
	Game
	// Connections is every live socket at the table, earliest first: []
	// for a table with none, absent only when the server does not know
	// its sockets (no hub).
	Connections *[]Connection `json:"connections,omitempty"`
}

// NewGameResponse is GET /admin/games/{id}'s answer: g, with the
// table's connections when the overlay knows its sockets.
func NewGameResponse(g Game, ov Overlay, now time.Time) GameResponse {
	out := GameResponse{GeneratedAt: now.UnixMilli(), Game: g}
	if id, err := uuid.Parse(g.ID); err == nil {
		if conns := Connections(id, ov); conns != nil {
			out.Connections = &conns
		}
	}
	return out
}

// Connection is one live socket at a table (§3.3).
type Connection struct {
	// Kind is seat, spectator or admin. An admin bound to a seat is an
	// admin, with that seat.
	Kind string `json:"kind"`
	// Seat is the seat the socket is bound to, if any.
	Seat *int `json:"seat,omitempty"`
	// Account is absent for a guest and for the shared token.
	Account *AccountRef `json:"account,omitempty"`
	// Since is when the hub admitted the connection.
	Since int64 `json:"since,omitempty"`
}

// The connection kinds, as GET /admin/live and cmdctrl_ws_connections
// {role} name them.
const (
	ConnSeat      = "seat"
	ConnSpectator = "spectator"
	ConnAdmin     = "admin"
)

// SocketUserIDs is the accounts behind the overlay's sockets at a
// table: the IDs a caller passes to Store.AccountRefs before
// Connections names them.
func (o Overlay) SocketUserIDs(gameID uuid.UUID) []string {
	var out []string
	for _, s := range o.Sockets {
		if s.GameID == gameID && s.UserID != "" {
			out = append(out, s.UserID)
		}
	}
	return out
}

// Connections is the table's live sockets, earliest first, for GET
// /admin/games/{id}; nil when the overlay does not know its sockets. An
// account's name and avatar come from o.Accounts, and a seated socket
// whose account has no row there takes the seat's name.
func Connections(gameID uuid.UUID, ov Overlay) []Connection {
	if !ov.SocketsKnown {
		return nil
	}
	t, _ := ov.table(gameID)
	seats := map[uuid.UUID]LiveSeat{}
	for _, s := range t.Seats {
		seats[s.PlayerID] = s
	}
	var sockets []Socket
	for _, s := range ov.Sockets {
		if s.GameID == gameID {
			sockets = append(sockets, s)
		}
	}
	sort.SliceStable(sockets, func(i, j int) bool { return sockets[i].ConnectedAt.Before(sockets[j].ConnectedAt) })
	out := make([]Connection, 0, len(sockets))
	for _, s := range sockets {
		c := Connection{Kind: ConnSpectator, Since: ms(s.ConnectedAt)}
		seat, seated := seats[s.PlayerID]
		if seated && s.PlayerID != uuid.Nil {
			n := seat.Seat
			c.Seat = &n
			c.Kind = ConnSeat
		}
		if s.admin() {
			c.Kind = ConnAdmin
		}
		if s.UserID != "" {
			ref := AccountRef{ID: s.UserID}
			if seated && seat.UserID == s.UserID {
				ref.Name = cmp.Or(seat.DisplayName, seat.Name)
			}
			if known, ok := ov.Accounts[s.UserID]; ok {
				ref = known
			}
			c.Account = &ref
		}
		out = append(out, c)
	}
	return out
}

// The seat kinds, as cmdctrl_seats{kind} names them.
const (
	KindHuman = "human"
	KindBot   = "bot"
	KindAgent = "agent"
)

// The outcomes a table row serves. A stored value this binary does
// not know is carried as it is.
const (
	OutcomeWin    = "win"
	OutcomeDraw   = "draw"
	OutcomeClosed = "closed"
)

// --- merging ---------------------------------------------------------

// MergeAccounts is GET /admin/users' answer: the rows, each marked
// playing_now when it is on live (Lobby.MetricsLiveUsers), in
// AccountsLess order.
func MergeAccounts(res AccountsResult, live []string, now time.Time) AccountsResponse {
	playing := set(live)
	out := AccountsResponse{GeneratedAt: now.UnixMilli(), Accounts: make([]Account, 0, len(res.Rows)), Truncated: res.Truncated}
	for _, r := range res.Rows {
		out.Accounts = append(out.Accounts, account(r, playing[r.ID]))
	}
	sort.SliceStable(out.Accounts, func(i, j int) bool { return AccountsLess(out.Accounts[i], out.Accounts[j]) })
	return out
}

// AccountsLess is the accounts list's order (§3.1): playing now first,
// then the latest last play, then the latest sign-in, newest first.
// The store's ORDER BY is the same, so its row cap keeps these rows.
func AccountsLess(a, b Account) bool {
	if a.PlayingNow != b.PlayingNow {
		return a.PlayingNow
	}
	if a.LastPlayedAt != b.LastPlayedAt {
		return a.LastPlayedAt > b.LastPlayedAt
	}
	if a.LastSignInAt != b.LastSignInAt {
		return a.LastSignInAt > b.LastSignInAt
	}
	return a.ID < b.ID
}

// MergeAccount is GET /admin/users/{id}'s answer.
func MergeAccount(rows AccountRows, live []string, ov Overlay, now time.Time) AccountResponse {
	out := AccountResponse{
		GeneratedAt: now.UnixMilli(),
		Account:     account(rows.Account, set(live)[rows.Account.ID]),
		SignIn: SignIn{
			LastSignInAt:          ms(rows.Account.LastSignInAt),
			DiscordLinkedAt:       ms(rows.DiscordLinkedAt),
			SessionsInvalidBefore: ms(rows.SessionsInvalidBefore),
			RevokePath:            "/admin/users/" + rows.Account.ID + "/revoke-sessions",
		},
		Games:                 make([]Game, 0, len(rows.Games)),
		GamesTruncated:        rows.GamesTruncated,
		Decks:                 make([]Deck, 0, len(rows.Decks)),
		DeckRequests:          make([]DeckRequest, 0, len(rows.DeckRequests)),
		DeckRequestsTruncated: rows.DeckRequestsTruncated,
	}
	for _, g := range rows.Games {
		out.Games = append(out.Games, MergeGame(g, ov))
	}
	for _, d := range rows.Decks {
		commanders := d.Commanders
		if commanders == nil {
			commanders = []string{}
		}
		out.Decks = append(out.Decks, Deck{
			ID: d.ID, Name: d.Name, Format: d.Format, SourceURL: d.SourceURL, Commanders: commanders,
			CardCount: d.CardCount, CreatedAt: ms(d.CreatedAt), UpdatedAt: ms(d.UpdatedAt),
		})
	}
	for _, d := range rows.DeckRequests {
		out.DeckRequests = append(out.DeckRequests, DeckRequest{
			DeckKey: d.DeckKey, AskedAt: ms(d.AskedAt), IssueNumber: d.IssueNumber, IssueURL: d.IssueURL,
		})
	}
	return out
}

// MergeGames is GET /admin/games' answer: the page's rows merged with
// memory, and the practice tables memory alone holds when the filter
// asks for them. practice=only lists only those; include puts them
// first on the first page.
func MergeGames(res GamesResult, ov Overlay, f GamesFilter, firstPage bool, now time.Time) GamesResponse {
	out := GamesResponse{GeneratedAt: now.UnixMilli(), Games: []Game{}}
	if f.Practice == PracticeOnly || (f.Practice == PracticeInclude && firstPage) {
		out.Games = append(out.Games, PracticeGames(ov, f)...)
	}
	if f.Practice == PracticeOnly {
		return out
	}
	for _, r := range res.Rows {
		out.Games = append(out.Games, MergeGame(r, ov))
	}
	if res.Next != nil {
		out.NextCursor = res.Next.String()
	}
	return out
}

// PracticeGames is the overlay's practice tables that pass the filter,
// newest first.
func PracticeGames(ov Overlay, f GamesFilter) []Game {
	var tables []LiveTable
	for _, t := range ov.Tables {
		if t.Practice && f.MatchesLive(t) {
			tables = append(tables, t)
		}
	}
	sort.SliceStable(tables, func(i, j int) bool {
		if !tables[i].CreatedAt.Equal(tables[j].CreatedAt) {
			return tables[i].CreatedAt.After(tables[j].CreatedAt)
		}
		return tables[i].ID.String() > tables[j].ID.String()
	})
	out := make([]Game, 0, len(tables))
	for _, t := range tables {
		out = append(out, LiveGame(t, ov))
	}
	return out
}

// MatchesLive applies the state, archived and user filters to a table
// only memory holds. The practice filter is the caller's.
func (f GamesFilter) MatchesLive(t LiveTable) bool {
	if f.State != "" && f.State != t.State {
		return false
	}
	if f.Archived != nil && *f.Archived != t.Archived {
		return false
	}
	if f.UserID != "" {
		for _, s := range t.Seats {
			if s.UserID == f.UserID {
				return true
			}
		}
		return false
	}
	return true
}

// MergeGame is one table row from its games row, with memory on top
// when the table is loaded: the cached state, the seat kinds, agent
// clients and host flags, and (when known) the connected counts.
func MergeGame(r GameRow, ov Overlay) Game {
	g := Game{
		ID:         r.ID.String(),
		Name:       r.Name,
		State:      r.State,
		CreatedAt:  ms(r.CreatedAt),
		StartedAt:  ms(r.StartedAt),
		EndedAt:    ms(r.EndedAt),
		ArchivedAt: ms(r.ArchivedAt),
		WinnerSeat: r.WinnerSeat,
		TheirSeat:  r.TheirSeat,
		Seats:      make([]Seat, 0, len(r.Seats)),
	}
	if r.CreatorID != "" {
		g.Creator = &Creator{ID: r.CreatorID, Name: r.CreatorName}
	}
	live, loaded := ov.table(r.ID)
	byPlayer := map[string]LiveSeat{}
	if loaded {
		g.Loaded = true
		g.Practice = live.Practice
		if live.State != "" {
			g.State = live.State
		}
		g.ArchivedAt = archivedAt(g.ArchivedAt, live.Archived)
		for _, s := range live.Seats {
			byPlayer[s.PlayerID.String()] = s
		}
	}
	seen := map[string]bool{}
	for _, s := range r.Seats {
		seat := rowSeat(s, r.HostPlayerID)
		if ls, ok := byPlayer[s.PlayerID]; ok {
			seen[s.PlayerID] = true
			overlaySeat(&seat, ls)
		}
		g.Seats = append(g.Seats, seat)
	}
	if loaded {
		// A seat memory holds and the row does not yet: shown from
		// memory, so the row never hides somebody at the table.
		for _, ls := range live.Seats {
			if !seen[ls.PlayerID.String()] {
				g.Seats = append(g.Seats, liveSeat(ls, ov.Accounts))
			}
		}
		sort.SliceStable(g.Seats, func(i, j int) bool { return g.Seats[i].Seat < g.Seats[j].Seat })
	}
	g.Outcome = outcome(r.Outcome, g.EndedAt != 0, r.StartedAt, r.ArchivedAt)
	connected(&g, live, loaded, ov)
	return g
}

// LiveGame is a table row from memory alone: a practice table, or a
// loaded table whose row the store does not have.
func LiveGame(t LiveTable, ov Overlay) Game {
	g := Game{
		ID:        t.ID.String(),
		Name:      t.Name,
		State:     t.State,
		CreatedAt: ms(t.CreatedAt),
		Practice:  t.Practice,
		Loaded:    true,
		Seats:     make([]Seat, 0, len(t.Seats)),
	}
	for _, s := range t.Seats {
		g.Seats = append(g.Seats, liveSeat(s, ov.Accounts))
	}
	sort.SliceStable(g.Seats, func(i, j int) bool { return g.Seats[i].Seat < g.Seats[j].Seat })
	connected(&g, t, true, ov)
	return g
}

// rowSeat is a seat from its row alone.
func rowSeat(s SeatRow, hostPlayerID string) Seat {
	seat := Seat{Seat: s.Seat, Kind: KindHuman, DeckName: s.DeckName, Host: hostPlayerID != "" && s.PlayerID == hostPlayerID}
	switch {
	case s.BotTier != "":
		seat.Kind, seat.BotTier = KindBot, s.BotTier
	case s.AgentClient != "":
		seat.Kind, seat.AgentClient = KindAgent, s.AgentClient
	}
	switch {
	case s.UserID != "":
		seat.Account = &AccountRef{ID: s.UserID, Name: s.UserName, AvatarURL: s.UserAvatarURL}
	default:
		seat.GuestName = s.GuestName
		seat.DiscordPending = s.DiscordPending
	}
	return seat
}

// overlaySeat puts memory's view of a loaded seat on top of its row:
// memory is the authority for the kind, the agent badge and the host.
func overlaySeat(seat *Seat, ls LiveSeat) {
	seat.Kind = ls.Kind
	seat.Host = ls.Host
	switch ls.Kind {
	case KindBot:
		if ls.BotTier != "" {
			seat.BotTier = ls.BotTier
		}
	case KindAgent:
		if ls.AgentClient != "" {
			seat.AgentClient = ls.AgentClient
		}
	}
	if seat.DeckName == "" {
		seat.DeckName = ls.DeckName
	}
}

// liveSeat is a seat from memory alone.
func liveSeat(ls LiveSeat, accounts map[string]AccountRef) Seat {
	seat := Seat{Seat: ls.Seat, Kind: ls.Kind, DeckName: ls.DeckName, Host: ls.Host}
	if seat.Kind == "" {
		seat.Kind = KindHuman
	}
	if ls.BotTier != "" {
		seat.BotTier = ls.BotTier
	}
	if ls.AgentClient != "" {
		seat.AgentClient = ls.AgentClient
	}
	name := ls.DisplayName
	if name == "" {
		name = ls.Name
	}
	if ls.UserID != "" {
		ref := AccountRef{ID: ls.UserID, Name: name}
		if known, ok := accounts[ls.UserID]; ok {
			ref = known
		}
		seat.Account = &ref
	} else {
		seat.GuestName = name
		seat.DiscordPending = ls.DiscordPending
	}
	return seat
}

// connected fills the connected counts of a loaded table when the
// overlay knows its sockets, by GET /admin/live's rules: every socket
// bound to a seat counts as that seat's, an admin's included, and a
// spectator is a read-only socket that is not an admin's.
func connected(g *Game, t LiveTable, loaded bool, ov Overlay) {
	if !loaded || !ov.SocketsKnown {
		return
	}
	perSeat := map[uuid.UUID]int{}
	spectators := 0
	for _, s := range ov.Sockets {
		if s.GameID != t.ID {
			continue
		}
		if s.PlayerID != uuid.Nil {
			perSeat[s.PlayerID]++
		}
		if s.spectator() {
			spectators++
		}
	}
	g.SpectatorsConnected = &spectators
	bySeat := map[int]uuid.UUID{}
	for _, s := range t.Seats {
		bySeat[s.Seat] = s.PlayerID
	}
	for i := range g.Seats {
		n := perSeat[bySeat[g.Seats[i].Seat]]
		g.Seats[i].Connected = &n
	}
}

// outcome is the served outcome (§3.3): win or draw as stored; closed
// for a table that ended, or was archived after it started, with none
// recorded (migration 0007's NULL, as cmdctrl_games_ended_total
// {outcome="closed"} counts it); absent for a table that has not ended.
func outcome(stored string, ended bool, started, archived time.Time) string {
	if stored != "" {
		return stored
	}
	if ended || (!started.IsZero() && !archived.IsZero()) {
		return OutcomeClosed
	}
	return ""
}

// archivedAt keeps the row's archived time while memory agrees the
// table is archived. Memory is fresher: a table unarchived since the
// row was read is not archived.
func archivedAt(row int64, archived bool) int64 {
	if !archived {
		return 0
	}
	return row
}

func account(r AccountRow, playing bool) Account {
	return Account{
		ID:           r.ID,
		Name:         r.Name,
		AvatarURL:    r.AvatarURL,
		FirstSeenAt:  ms(r.FirstSeenAt),
		LastSignInAt: ms(r.LastSignInAt),
		GamesPlayed:  r.GamesPlayed,
		LastPlayedAt: ms(r.LastPlayedAt),
		PlayingNow:   playing,
	}
}

func set(ids []string) map[string]bool {
	out := make(map[string]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out
}

// ms is t in Unix milliseconds, 0 for the zero time (absent on the
// wire).
func ms(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixMilli()
}
