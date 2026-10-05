package adminview

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

// --- a seeded database ---------------------------------------------

type seeded struct {
	t *testing.T
	d *db.DB
}

func openSeeded(t *testing.T) *seeded {
	t.Helper()
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return &seeded{t: t, d: d}
}

func (s *seeded) exec(q string, args ...any) {
	s.t.Helper()
	if _, err := s.d.Exec(q, args...); err != nil {
		s.t.Fatalf("seed %q: %v", q, err)
	}
}

func (s *seeded) user(name string, created, seen time.Time) string {
	s.t.Helper()
	id := uuid.NewString()
	s.exec(`INSERT INTO users (id, display_name, avatar_url, created_at, last_seen_at) VALUES (?, ?, ?, ?, ?)`,
		id, name, "/avatars/123/abc.png", created.UnixMilli(), seen.UnixMilli())
	return id
}

func (s *seeded) identity(userID, subject string, linked time.Time) {
	s.t.Helper()
	s.exec(`INSERT INTO identities (provider, subject, user_id, display_name, refresh_token, scopes, linked_at)
	        VALUES ('discord', ?, ?, 'x', X'00112233', 'identify', ?)`, subject, userID, linked.UnixMilli())
}

// game is one games row. A zero time is NULL.
type game struct {
	id                                  uuid.UUID
	name, state, outcome, host, creator string
	created, started, ended, archived   time.Time
	winner                              *int
}

func nullable(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.UnixMilli()
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (s *seeded) game(g game) uuid.UUID {
	s.t.Helper()
	if g.id == uuid.Nil {
		g.id = uuid.New()
	}
	if g.state == "" {
		g.state = "lobby"
	}
	if g.name == "" {
		g.name = "table"
	}
	var winner any
	if g.winner != nil {
		winner = *g.winner
	}
	s.exec(`INSERT INTO games (id, name, created_by, state, created_at, started_at, ended_at, archived_at, winner_seat, host_player_id, outcome)
	        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		g.id.String(), g.name, nullStr(g.creator), g.state, g.created.UnixMilli(), nullable(g.started), nullable(g.ended),
		nullable(g.archived), winner, nullStr(g.host), nullStr(g.outcome))
	return g.id
}

// seat is one seats row; empty strings are NULL.
type seat struct {
	n                                              int
	player, user, guest, bot, agent, deck, discord string
}

func (s *seeded) seat(gameID uuid.UUID, st seat) string {
	s.t.Helper()
	if st.player == "" {
		st.player = uuid.NewString()
	}
	s.exec(`INSERT INTO seats (game_id, seat, player_id, user_id, guest_name, bot_tier, agent_client, deck_name, pending_discord_id)
	        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		gameID.String(), st.n, st.player, nullStr(st.user), nullStr(st.guest), nullStr(st.bot), nullStr(st.agent),
		nullStr(st.deck), nullStr(st.discord))
	return st.player
}

var t0 = time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

func at(d time.Duration) time.Time { return t0.Add(d) }

func accountIDs(rows []AccountRow) []string {
	var out []string
	for _, r := range rows {
		out = append(out, r.ID)
	}
	return out
}

// --- accounts --------------------------------------------------------

func TestAccountsCountStartedTablesAndTheirLatestEnd(t *testing.T) {
	s := openSeeded(t)
	ann := s.user("Ann", t0, at(time.Hour))
	bea := s.user("Bea", t0, at(2*time.Hour))

	ended := s.game(game{state: "ended", created: at(1 * time.Minute), started: at(2 * time.Minute), ended: at(10 * time.Hour)})
	archivedRunning := s.game(game{state: "active", created: at(3 * time.Minute), started: at(4 * time.Minute), archived: at(20 * time.Hour)})
	neverStarted := s.game(game{state: "lobby", created: at(5 * time.Minute), archived: at(30 * time.Hour)})
	running := s.game(game{state: "active", created: at(6 * time.Minute), started: at(7 * time.Minute)})
	for _, g := range []uuid.UUID{ended, archivedRunning, neverStarted, running} {
		s.seat(g, seat{n: 0, user: ann})
	}

	res, err := NewSQLStore(s.d).Accounts(context.Background(), AccountsQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) != 2 || res.Truncated {
		t.Fatalf("rows = %+v truncated %v, want two rows", res.Rows, res.Truncated)
	}
	got := map[string]AccountRow{}
	for _, r := range res.Rows {
		got[r.ID] = r
	}
	if a := got[ann]; a.GamesPlayed != 3 || !a.LastPlayedAt.Equal(at(20*time.Hour)) {
		t.Errorf("Ann: games_played %d, last played %v; want 3 (started tables only) and the archived table's end", a.GamesPlayed, a.LastPlayedAt)
	}
	if a := got[ann]; a.Name != "Ann" || !a.FirstSeenAt.Equal(t0) || !a.LastSignInAt.Equal(at(time.Hour)) || a.AvatarURL == "" {
		t.Errorf("Ann's row = %+v", a)
	}
	if b := got[bea]; b.GamesPlayed != 0 || !b.LastPlayedAt.IsZero() {
		t.Errorf("Bea: games_played %d, last played %v; want none", b.GamesPlayed, b.LastPlayedAt)
	}
}

func TestAccountsSortPlayingNowFirstAndTruncate(t *testing.T) {
	s := openSeeded(t)
	early := s.user("Early", t0, at(time.Hour))
	late := s.user("Late", t0, at(3*time.Hour))
	player := s.user("Player", t0, at(2*time.Hour))
	live := s.user("Live", t0, t0)

	g := s.game(game{state: "ended", created: at(1 * time.Minute), started: at(2 * time.Minute), ended: at(5 * time.Hour)})
	s.seat(g, seat{n: 0, user: player})

	store := NewSQLStore(s.d)
	res, err := store.Accounts(context.Background(), AccountsQuery{Live: []string{live, live}})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{live, player, late, early}
	if got := accountIDs(res.Rows); !slices.Equal(got, want) {
		t.Errorf("order = %v, want playing now, last played, then last sign-in: %v", got, want)
	}

	res, err = store.Accounts(context.Background(), AccountsQuery{Live: []string{live}, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if got := accountIDs(res.Rows); !slices.Equal(got, want[:2]) || !res.Truncated {
		t.Errorf("limit 2: %v truncated %v, want %v and truncated", got, res.Truncated, want[:2])
	}
}

// TestPlayedFilterAgreesWithTheGauge is ADR 0124 §8: played=<window>
// keeps exactly the accounts cmdctrl_users_played{window} counts, the
// ones users.SQLStore.LastPlayed returns for the window's start plus
// the accounts playing now, over the gauge's own windows.
func TestPlayedFilterAgreesWithTheGauge(t *testing.T) {
	s := openSeeded(t)
	now := t0.Add(40 * 24 * time.Hour)
	day := 24 * time.Hour

	// Ends spread across and around every window's edge.
	offsets := []time.Duration{
		0, time.Hour, day - time.Millisecond, day, day + time.Millisecond, 3 * day,
		7*day - time.Millisecond, 7 * day, 7*day + time.Millisecond, 20 * day,
		30 * day, 30*day + time.Millisecond, 35 * day,
	}
	for i, off := range offsets {
		u := s.user(fmt.Sprintf("U%d", i), t0, t0)
		end := now.Add(-off)
		var g uuid.UUID
		if i%2 == 0 {
			g = s.game(game{state: "ended", created: t0, started: t0, ended: end})
		} else {
			// Closed by archiving while it ran.
			g = s.game(game{state: "active", created: t0, started: t0, archived: end})
		}
		s.seat(g, seat{n: 0, user: u})
	}
	// Seated at a table that never started, archived inside every
	// window: not a play.
	idle := s.user("Idle", t0, now)
	g := s.game(game{state: "lobby", created: t0, archived: now.Add(-time.Hour)})
	s.seat(g, seat{n: 0, user: idle})
	// Playing now, with no ended table at all.
	live := s.user("Live", t0, t0)
	// Nothing at all.
	s.user("Never", t0, t0)

	store := NewSQLStore(s.d)
	gauge := users.NewSQLStore(s.d, nil)
	for _, label := range []string{"1d", "7d", "30d"} {
		span, ok := metrics.UsersPlayedWindow(label)
		if !ok {
			t.Fatalf("metrics has no window %q", label)
		}
		since := now.Add(-span)
		ends, err := gauge.LastPlayed(context.Background(), since)
		if err != nil {
			t.Fatal(err)
		}
		want := []string{live}
		for id := range ends {
			want = append(want, id)
		}
		sort.Strings(want)

		res, err := store.Accounts(context.Background(), AccountsQuery{PlayedSince: since, Live: []string{live}})
		if err != nil {
			t.Fatal(err)
		}
		got := accountIDs(res.Rows)
		sort.Strings(got)
		if !slices.Equal(got, want) {
			t.Errorf("played=%s: the page keeps %v, the gauge counts %v", label, got, want)
		}
		if len(got) < 2 {
			t.Errorf("played=%s keeps only %d accounts; the fixture is not exercising the window", label, len(got))
		}
	}
	if _, ok := metrics.UsersPlayedWindow("2d"); ok {
		t.Error("metrics reports a 2d window")
	}
}

// --- one account -----------------------------------------------------

func TestAccountReadsSignInDecksAndDeckRequests(t *testing.T) {
	s := openSeeded(t)
	ann := s.user("Ann", t0, at(time.Hour))
	s.identity(ann, "4242", at(time.Minute))
	s.exec(`UPDATE users SET sessions_invalid_before = ? WHERE id = ?`, at(2*time.Hour).UnixMilli(), ann)
	bea := s.user("Bea", t0, t0)
	s.identity(bea, "5151", t0)

	s.exec(`INSERT INTO decks (id, owner_id, name, source_format, source_text, commanders, card_count, created_at, updated_at, source_url)
	        VALUES ('d1', ?, 'Old', 'text', '1 Sol Ring', '["Atraxa"]', 100, ?, ?, NULL),
	               ('d2', ?, 'New', 'moxfield', '1 Sol Ring', '["Kenrith","Partner"]', 100, ?, ?, 'https://moxfield.com/decks/x')`,
		ann, at(1*time.Minute).UnixMilli(), at(1*time.Minute).UnixMilli(), ann, at(2*time.Minute).UnixMilli(), at(3*time.Minute).UnixMilli())

	s.exec(`INSERT INTO deck_requests (deck_key, issue_number, issue_url, created_at) VALUES ('moxfield:a', 7, 'https://github.com/o/r/issues/7', 1)`)
	s.exec(`INSERT INTO deck_request_asks (deck_key, requester, at) VALUES
	        ('moxfield:a', 'discord:4242', ?), ('archidekt:b', 'discord:4242', ?), ('moxfield:a', 'discord:5151', ?), ('moxfield:a', 'discord:999', ?)`,
		at(10*time.Minute).UnixMilli(), at(20*time.Minute).UnixMilli(), at(30*time.Minute).UnixMilli(), at(40*time.Minute).UnixMilli())

	rows, err := NewSQLStore(s.d).Account(context.Background(), uuid.MustParse(ann))
	if err != nil {
		t.Fatal(err)
	}
	if rows.Account.ID != ann || rows.Account.Name != "Ann" {
		t.Errorf("account = %+v", rows.Account)
	}
	if !rows.DiscordLinkedAt.Equal(at(time.Minute)) || !rows.SessionsInvalidBefore.Equal(at(2*time.Hour)) {
		t.Errorf("sign-in: linked %v, invalid before %v", rows.DiscordLinkedAt, rows.SessionsInvalidBefore)
	}
	if len(rows.Decks) != 2 || rows.Decks[0].ID != "d2" || rows.Decks[0].SourceURL == "" ||
		!slices.Equal(rows.Decks[0].Commanders, []string{"Kenrith", "Partner"}) || rows.Decks[1].SourceURL != "" {
		t.Errorf("decks = %+v, want New then Old", rows.Decks)
	}
	want := []DeckRequestRow{
		{DeckKey: "archidekt:b", AskedAt: millis(at(20 * time.Minute).UnixMilli())},
		{DeckKey: "moxfield:a", AskedAt: millis(at(10 * time.Minute).UnixMilli()), IssueNumber: 7, IssueURL: "https://github.com/o/r/issues/7"},
	}
	if !slices.Equal(rows.DeckRequests, want) || rows.DeckRequestsTruncated {
		t.Errorf("deck requests = %+v, want Ann's two asks through her identity, newest first: %+v", rows.DeckRequests, want)
	}

	// No identity, no sessions revoked.
	cleo := s.user("Cleo", t0, t0)
	rows, err = NewSQLStore(s.d).Account(context.Background(), uuid.MustParse(cleo))
	if err != nil {
		t.Fatal(err)
	}
	if !rows.DiscordLinkedAt.IsZero() || !rows.SessionsInvalidBefore.IsZero() || len(rows.DeckRequests) != 0 || len(rows.Decks) != 0 || len(rows.Games) != 0 {
		t.Errorf("an account with nothing = %+v", rows)
	}

	if _, err := NewSQLStore(s.d).Account(context.Background(), uuid.New()); err != ErrNotFound {
		t.Errorf("an unknown account: %v, want ErrNotFound", err)
	}
}

// countingQuerier counts the store's queries.
type countingQuerier struct {
	q Querier
	n atomic.Int64
}

func (c *countingQuerier) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	c.n.Add(1)
	return c.q.QueryContext(ctx, query, args...)
}

func (c *countingQuerier) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	c.n.Add(1)
	return c.q.QueryRowContext(ctx, query, args...)
}

// An account's games are two queries, whatever their number, and so
// is a page of games: no query per row (ADR 0124 §3.2, §3.3).
func TestQueriesDoNotGrowWithTheRows(t *testing.T) {
	counts := map[int][3]int64{}
	for _, n := range []int{1, 9} {
		s := openSeeded(t)
		ann := s.user("Ann", t0, t0)
		for i := range n {
			g := s.game(game{state: "ended", created: at(time.Duration(i) * time.Minute), started: at(1 * time.Minute), ended: at(2 * time.Minute)})
			s.seat(g, seat{n: 0, user: ann})
			s.seat(g, seat{n: 1, guest: "Guest"})
			s.seat(g, seat{n: 2, bot: "random"})
		}
		cq := &countingQuerier{q: s.d}
		store := NewQuerierStore(cq)
		rows, err := store.Account(context.Background(), uuid.MustParse(ann))
		if err != nil {
			t.Fatal(err)
		}
		if len(rows.Games) != n || len(rows.Games[0].Seats) != 3 {
			t.Fatalf("%d games: got %d games, first with %d seats", n, len(rows.Games), len(rows.Games[0].Seats))
		}
		account := cq.n.Swap(0)
		page, err := store.Games(context.Background(), GamesQuery{})
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Rows) != n {
			t.Fatalf("games page has %d rows, want %d", len(page.Rows), n)
		}
		games := cq.n.Swap(0)
		if _, err := store.Accounts(context.Background(), AccountsQuery{}); err != nil {
			t.Fatal(err)
		}
		counts[n] = [3]int64{account, games, cq.n.Load()}
	}
	if counts[1] != counts[9] {
		t.Errorf("queries for 1 game %v, for 9 games %v; a query per row crept in", counts[1], counts[9])
	}
	if c := counts[9]; c[0] != 5 || c[1] != 2 || c[2] != 1 {
		t.Errorf("queries (account, games page, accounts) = %v, want 5 (row, games, seats, decks, deck requests), 2 and 1", c)
	}
}

// --- games -----------------------------------------------------------

// Keyset pages never repeat or skip a table, including tables created
// in the same millisecond.
func TestGamesPagesNeverRepeatOrSkip(t *testing.T) {
	s := openSeeded(t)
	var want []uuid.UUID
	for i := range 11 {
		created := at(time.Duration(i/4) * time.Second) // runs of four share a created_at
		want = append(want, s.game(game{created: created}))
	}
	store := NewSQLStore(s.d)
	full, err := store.Games(context.Background(), GamesQuery{Limit: MaxGamesLimit})
	if err != nil {
		t.Fatal(err)
	}
	if full.Next != nil || len(full.Rows) != len(want) {
		t.Fatalf("one big page: %d rows, next %v", len(full.Rows), full.Next)
	}
	for i := 1; i < len(full.Rows); i++ {
		a, b := full.Rows[i-1], full.Rows[i]
		if a.CreatedAt.Before(b.CreatedAt) || (a.CreatedAt.Equal(b.CreatedAt) && a.ID.String() < b.ID.String()) {
			t.Fatalf("not newest first at %d: %v %s then %v %s", i, a.CreatedAt, a.ID, b.CreatedAt, b.ID)
		}
	}

	for _, limit := range []int{1, 2, 3, 4, 5} {
		var got []uuid.UUID
		var after *Cursor
		for pages := 0; ; pages++ {
			if pages > len(want) {
				t.Fatalf("limit %d: more pages than tables", limit)
			}
			page, err := store.Games(context.Background(), GamesQuery{Limit: limit, After: after})
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range page.Rows {
				got = append(got, r.ID)
			}
			if page.Next == nil {
				break
			}
			// Round-trip the cursor as a client would.
			c, err := ParseCursor(page.Next.String())
			if err != nil {
				t.Fatal(err)
			}
			after = &c
		}
		var wantOrder []uuid.UUID
		for _, r := range full.Rows {
			wantOrder = append(wantOrder, r.ID)
		}
		if !slices.Equal(got, wantOrder) {
			t.Errorf("limit %d: pages gave %v, want %v", limit, got, wantOrder)
		}
	}
}

func TestGamesFiltersAreInSQL(t *testing.T) {
	s := openSeeded(t)
	ann := s.user("Ann", t0, t0)
	lobby := s.game(game{state: "lobby", created: at(1 * time.Minute)})
	active := s.game(game{state: "active", created: at(2 * time.Minute), started: at(2 * time.Minute)})
	archived := s.game(game{state: "active", created: at(3 * time.Minute), started: at(3 * time.Minute), archived: at(4 * time.Minute)})
	ended := s.game(game{state: "ended", created: at(4 * time.Minute), started: at(4 * time.Minute), ended: at(5 * time.Minute)})
	s.seat(active, seat{n: 0, user: ann})
	s.seat(ended, seat{n: 1, user: ann})
	s.seat(lobby, seat{n: 0, guest: "Guest"})

	yes, no := true, false
	cases := []struct {
		name string
		q    GamesQuery
		want []uuid.UUID
	}{
		{"any", GamesQuery{}, []uuid.UUID{ended, archived, active, lobby}},
		{"state=active", GamesQuery{State: "active"}, []uuid.UUID{archived, active}},
		{"state=active&archived=false", GamesQuery{State: "active", Archived: &no}, []uuid.UUID{active}},
		{"archived=true", GamesQuery{Archived: &yes}, []uuid.UUID{archived}},
		{"user", GamesQuery{UserID: ann}, []uuid.UUID{ended, active}},
		{"state=lobby&user", GamesQuery{State: "lobby", UserID: ann}, nil},
	}
	store := NewSQLStore(s.d)
	for _, tc := range cases {
		res, err := store.Games(context.Background(), tc.q)
		if err != nil {
			t.Fatal(err)
		}
		var got []uuid.UUID
		for _, r := range res.Rows {
			got = append(got, r.ID)
		}
		if !slices.Equal(got, tc.want) {
			t.Errorf("%s: %v, want %v", tc.name, got, tc.want)
		}
	}
}

// Each seat kind maps right from its row alone: an account, a guest, a
// pending Discord seat, a bot, and an agent from migration 0010's
// column.
func TestSeatKindsFromTheirRows(t *testing.T) {
	s := openSeeded(t)
	ann := s.user("Ann", t0, t0)
	creator := s.user("Creator", t0, t0)
	host := uuid.NewString()
	two := 2
	g := s.game(game{state: "ended", created: at(1 * time.Minute), started: at(1 * time.Minute), ended: at(2 * time.Minute), outcome: "win", winner: &two, host: host, creator: creator})
	s.seat(g, seat{n: 0, player: host, user: ann, deck: "Ann's deck"})
	s.seat(g, seat{n: 1, guest: "Gus"})
	s.seat(g, seat{n: 2, guest: "Dee", discord: "777"})
	s.seat(g, seat{n: 3, guest: "Bot 1", bot: "heuristic"})
	s.seat(g, seat{n: 4, guest: "Claude", agent: "claude-code"})

	row, err := NewSQLStore(s.d).Game(context.Background(), g)
	if err != nil {
		t.Fatal(err)
	}
	got := MergeGame(row, Overlay{})
	if got.Loaded || got.Practice || got.Outcome != OutcomeWin || got.WinnerSeat == nil || *got.WinnerSeat != 2 {
		t.Errorf("game = %+v", got)
	}
	if got.Creator == nil || got.Creator.ID != creator || got.Creator.Name != "Creator" {
		t.Errorf("creator = %+v", got.Creator)
	}
	want := []Seat{
		{Seat: 0, Kind: KindHuman, Account: &AccountRef{ID: ann, Name: "Ann", AvatarURL: "/avatars/123/abc.png"}, DeckName: "Ann's deck", Host: true},
		{Seat: 1, Kind: KindHuman, GuestName: "Gus"},
		{Seat: 2, Kind: KindHuman, GuestName: "Dee", DiscordPending: true},
		{Seat: 3, Kind: KindBot, GuestName: "Bot 1", BotTier: "heuristic"},
		{Seat: 4, Kind: KindAgent, GuestName: "Claude", AgentClient: "claude-code"},
	}
	if len(got.Seats) != len(want) {
		t.Fatalf("%d seats, want %d", len(got.Seats), len(want))
	}
	for i := range want {
		g, w := got.Seats[i], want[i]
		if g.Seat != w.Seat || g.Kind != w.Kind || g.GuestName != w.GuestName || g.DiscordPending != w.DiscordPending ||
			g.BotTier != w.BotTier || g.AgentClient != w.AgentClient || g.DeckName != w.DeckName || g.Host != w.Host ||
			(g.Account == nil) != (w.Account == nil) || (g.Account != nil && *g.Account != *w.Account) || g.Connected != nil {
			t.Errorf("seat %d = %+v (account %+v), want %+v (account %+v)", i, g, g.Account, w, w.Account)
		}
	}

	if _, err := NewSQLStore(s.d).Game(context.Background(), uuid.New()); err != ErrNotFound {
		t.Errorf("an unknown game: %v, want ErrNotFound", err)
	}
}

func TestOutcomeFromTheRows(t *testing.T) {
	s := openSeeded(t)
	cases := []struct {
		name string
		g    game
		want string
	}{
		{"a win", game{state: "ended", started: at(1 * time.Minute), ended: at(2 * time.Minute), outcome: "win"}, OutcomeWin},
		{"a draw", game{state: "ended", started: at(1 * time.Minute), ended: at(2 * time.Minute), outcome: "draw"}, OutcomeDraw},
		{"ended with none recorded", game{state: "ended", started: at(1 * time.Minute), ended: at(2 * time.Minute)}, OutcomeClosed},
		{"archived while it ran", game{state: "active", started: at(1 * time.Minute), archived: at(2 * time.Minute)}, OutcomeClosed},
		{"archived before it started", game{state: "lobby", archived: at(2 * time.Minute)}, ""},
		{"running", game{state: "active", started: at(1 * time.Minute)}, ""},
		{"waiting", game{state: "lobby"}, ""},
	}
	store := NewSQLStore(s.d)
	for _, tc := range cases {
		tc.g.created = t0
		id := s.game(tc.g)
		row, err := store.Game(context.Background(), id)
		if err != nil {
			t.Fatal(err)
		}
		if got := MergeGame(row, Overlay{}).Outcome; got != tc.want {
			t.Errorf("%s: outcome %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestAccountRefsNamesOnlyKnownAccounts(t *testing.T) {
	s := openSeeded(t)
	ann := s.user("Ann", t0, t0)
	refs, err := NewSQLStore(s.d).AccountRefs(context.Background(), []string{ann, uuid.NewString()})
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 1 || refs[ann].Name != "Ann" || refs[ann].AvatarURL == "" {
		t.Errorf("refs = %+v", refs)
	}
	refs, err = NewSQLStore(s.d).AccountRefs(context.Background(), nil)
	if err != nil || len(refs) != 0 {
		t.Errorf("no ids: %v %v", refs, err)
	}
}
