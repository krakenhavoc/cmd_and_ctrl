// Package adminview is the read side of the admin views (ADR 0124 §5):
// every SQL statement behind GET /admin/users, /admin/users/{id},
// /admin/games and /admin/games/{id}, the rows they return, the JSON
// those routes answer, and the pure functions that merge the rows with
// what only memory knows (the lobby's loaded tables, practice tables,
// the accounts playing now).
//
// Three rules hold the package:
//
//   - It only reads. SQLStore holds a Querier, which has no Exec and no
//     transaction, and TestAdminViewStoreOnlyReads fails on a non-test
//     file that calls ExecContext or BeginTx, or passes a query that is
//     not a string literal starting with SELECT or WITH.
//   - It imports internal/db and uuid, and not lobby or ws. The live
//     overlay comes in as plain values (LiveTable, LiveSeat, Socket),
//     which lobby/admin_views.go copies from Lobby.LiveTables.
//   - The field lists of ADR 0124 §3 are the whole response. A field
//     added to a response type here is a deliberate edit to
//     lobby.TestAdminViewsServeOnlyTheirFields as well. No snowflake,
//     deck list, deck-request requester, token, scope or address is
//     ever selected into a row, so none can reach a response.
//
// Every time is Unix milliseconds on the wire, as on /me/games, and
// absent when unknown.
package adminview

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound is the store's answer for an account or a table with no
// row.
var ErrNotFound = errors.New("adminview: not found")

// The bounds of ADR 0124 §2.
const (
	// MaxAccounts is GET /admin/users' row cap; past it the answer says
	// truncated.
	MaxAccounts = 1000
	// MaxAccountGames and MaxAccountDeckRequests cap one account's
	// games and deck requests, each with its own truncated flag. Decks
	// need no cap here: the library holds at most 200 per person.
	MaxAccountGames        = 500
	MaxAccountDeckRequests = 100
	// DefaultGamesLimit and MaxGamesLimit are GET /admin/games' page
	// size.
	DefaultGamesLimit = 50
	MaxGamesLimit     = 200
)

// Store is the admin views' read-only store. SQLStore implements it
// over the server's database; a server with no database has none, and
// the four routes answer 503.
type Store interface {
	// Accounts is every account the query keeps, at most q.Limit, in
	// AccountsLess order.
	Accounts(ctx context.Context, q AccountsQuery) (AccountsResult, error)
	// Account is one account, with its games, decks and deck requests.
	// ErrNotFound when there is no such users row.
	Account(ctx context.Context, id uuid.UUID) (AccountRows, error)
	// Games is one page of tables, newest first.
	Games(ctx context.Context, q GamesQuery) (GamesResult, error)
	// Game is one table's row and seats. ErrNotFound when there is no
	// such games row.
	Game(ctx context.Context, id uuid.UUID) (GameRow, error)
	// AccountRefs names the accounts behind seats that have no row of
	// their own (a practice table's), in one query. An ID with no users
	// row is absent from the map.
	AccountRefs(ctx context.Context, ids []string) (map[string]AccountRef, error)
}

// --- queries and rows ------------------------------------------------

// AccountsQuery is GET /admin/users' query.
type AccountsQuery struct {
	// PlayedSince keeps an account whose last play ended at or after
	// it, or that is playing now. The zero time keeps every account.
	PlayedSince time.Time
	// Live is the accounts seated at a running table now
	// (Lobby.MetricsLiveUsers): kept by any PlayedSince, and sorted
	// first. Duplicates are harmless.
	Live []string
	// Limit is the row cap; <= 0 means MaxAccounts.
	Limit int
}

// AccountsResult is Accounts' answer.
type AccountsResult struct {
	Rows      []AccountRow
	Truncated bool
}

// AccountRow is one account as the database knows it.
type AccountRow struct {
	ID           string
	Name         string
	AvatarURL    string
	FirstSeenAt  time.Time // users.created_at
	LastSignInAt time.Time // users.last_seen_at
	// GamesPlayed is the distinct started tables the account holds a
	// seat at, in any state.
	GamesPlayed int
	// LastPlayedAt is the latest end (ended_at, else archived_at) of
	// those tables; zero when none has ended. The same rule as
	// users.SQLStore.LastPlayed, behind cmdctrl_users_played.
	LastPlayedAt time.Time
}

// AccountRows is Account's answer: everything GET /admin/users/{id}
// reads from the database.
type AccountRows struct {
	Account         AccountRow
	DiscordLinkedAt time.Time // identities.linked_at; zero with no identity
	// Playmats are the account's saved playmats, in slot order. The
	// handler fills them from the playmat service (ADR 0128 §11); they
	// are not columns the store reads.
	Playmats              []PlaymatSlot
	SessionsInvalidBefore time.Time // zero while users.sessions_invalid_before is 0
	Games                 []GameRow // newest first, each with TheirSeat
	GamesTruncated        bool
	Decks                 []DeckRow // newest first
	DeckRequests          []DeckRequestRow
	DeckRequestsTruncated bool
}

// DeckRow is one saved deck, without its list.
type DeckRow struct {
	ID         string
	Name       string
	Format     string
	SourceURL  string
	Commanders []string
	CardCount  int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// DeckRequestRow is one ask the account made, with the issue that
// tracks the deck now. The requester itself is never selected.
type DeckRequestRow struct {
	DeckKey     string
	AskedAt     time.Time
	IssueNumber int
	IssueURL    string
}

// GamesQuery is one page of GET /admin/games. Every filter is in SQL.
type GamesQuery struct {
	State    string // "" any, else lobby | active | ended
	Archived *bool  // nil any
	UserID   string // "" any, else tables this account holds a seat at
	Limit    int    // <= 0 means DefaultGamesLimit; capped at MaxGamesLimit
	After    *Cursor
}

// GamesResult is Games' answer. Next is set when there is more.
type GamesResult struct {
	Rows []GameRow
	Next *Cursor
}

// GameRow is one games row and its seats.
type GameRow struct {
	ID           uuid.UUID
	Name         string
	State        string
	CreatedAt    time.Time
	StartedAt    time.Time
	EndedAt      time.Time
	ArchivedAt   time.Time
	Outcome      string // games.outcome as stored, "" for NULL
	WinnerSeat   *int
	CreatorID    string
	CreatorName  string
	HostPlayerID string
	// TheirSeat is the seat the account holds, on an account's games.
	TheirSeat *int
	Seats     []SeatRow
}

// SeatRow is one seats row, with its account's name when it has one.
type SeatRow struct {
	Seat           int
	PlayerID       string
	UserID         string
	UserName       string
	UserAvatarURL  string
	GuestName      string
	DiscordPending bool // pending_discord_id is set; the snowflake is not selected
	BotTier        string
	AgentClient    string // seats.agent_client (migration 0010)
	DeckName       string
}

// --- the cursor ------------------------------------------------------

// Cursor is a keyset position in GET /admin/games: the last row of the
// page before, by (created_at, id), newest first.
type Cursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

// ErrBadCursor is ParseCursor's answer for a value this server did not
// issue.
var ErrBadCursor = errors.New("adminview: bad cursor")

// String is the cursor's opaque form, as next_cursor serves it.
func (c Cursor) String() string {
	raw := strconv.FormatInt(c.CreatedAt.UnixMilli(), 10) + "." + c.ID.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// ParseCursor reads a next_cursor value back.
func ParseCursor(s string) (Cursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, ErrBadCursor
	}
	ms, id, ok := strings.Cut(string(raw), ".")
	if !ok {
		return Cursor{}, ErrBadCursor
	}
	n, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return Cursor{}, ErrBadCursor
	}
	gid, err := uuid.Parse(id)
	if err != nil {
		return Cursor{}, ErrBadCursor
	}
	return Cursor{CreatedAt: time.UnixMilli(n), ID: gid}, nil
}

// --- filters ---------------------------------------------------------

// The closed values of GET /admin/games' practice filter.
const (
	PracticeExclude = "exclude"
	PracticeInclude = "include"
	PracticeOnly    = "only"
)

// GamesFilter is GET /admin/games' filters, applied in SQL to rows and
// by MatchesLive to the tables only memory holds.
type GamesFilter struct {
	State    string
	Archived *bool
	UserID   string
	Practice string // PracticeExclude, PracticeInclude or PracticeOnly
}

// ParseGamesFilter reads GET /admin/games' filters. A missing or empty
// parameter means any; practice defaults to exclude, so the unfiltered
// list counts what the Grafana tiles count. An unknown value is an
// error that names the parameter.
func ParseGamesFilter(get func(string) string) (GamesFilter, error) {
	f := GamesFilter{Practice: PracticeExclude}
	switch v := get("state"); v {
	case "", "lobby", "active", "ended":
		f.State = v
	default:
		return f, fmt.Errorf("state must be lobby, active or ended")
	}
	switch get("archived") {
	case "":
	case "true":
		t := true
		f.Archived = &t
	case "false":
		fl := false
		f.Archived = &fl
	default:
		return f, fmt.Errorf("archived must be true or false")
	}
	switch v := get("practice"); v {
	case "":
	case PracticeExclude, PracticeInclude, PracticeOnly:
		f.Practice = v
	default:
		return f, fmt.Errorf("practice must be exclude, include or only")
	}
	if v := get("user"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil || id == uuid.Nil {
			return f, fmt.Errorf("user must be an account id")
		}
		f.UserID = id.String()
	}
	return f, nil
}

// ParseGamesLimit reads GET /admin/games' limit: DefaultGamesLimit when
// missing, capped at MaxGamesLimit, and an error for anything that is
// not a positive whole number.
func ParseGamesLimit(v string) (int, error) {
	if v == "" {
		return DefaultGamesLimit, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("limit must be a whole number from 1 to %d", MaxGamesLimit)
	}
	return min(n, MaxGamesLimit), nil
}
