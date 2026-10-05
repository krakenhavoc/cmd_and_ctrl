package adminview

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMergeAccountsMarksPlayingNowAndSorts(t *testing.T) {
	res := AccountsResult{Rows: []AccountRow{
		{ID: "a", Name: "Old", LastSignInAt: at(time.Hour)},
		{ID: "b", Name: "Played", LastPlayedAt: at(2 * time.Hour), GamesPlayed: 3},
		{ID: "c", Name: "Live", LastSignInAt: t0},
	}, Truncated: true}
	now := at(10 * time.Hour)
	got := MergeAccounts(res, []string{"c"}, now)
	var order []string
	for _, a := range got.Accounts {
		order = append(order, a.ID)
	}
	if !slices.Equal(order, []string{"c", "b", "a"}) {
		t.Errorf("order %v, want playing now, then last played, then last sign-in", order)
	}
	if !got.Accounts[0].PlayingNow || got.Accounts[1].PlayingNow || !got.Truncated || got.GeneratedAt != now.UnixMilli() {
		t.Errorf("response = %+v", got)
	}
	if b := got.Accounts[1]; b.GamesPlayed != 3 || b.LastPlayedAt != at(2*time.Hour).UnixMilli() || b.FirstSeenAt != 0 {
		t.Errorf("Played = %+v; a zero time must be absent", b)
	}
	if empty := MergeAccounts(AccountsResult{}, nil, now); empty.Accounts == nil {
		t.Error("no rows serves null, want []")
	}
}

func TestMergeAccountServesItsSections(t *testing.T) {
	g := uuid.New()
	seat := 1
	rows := AccountRows{
		Account:               AccountRow{ID: "u1", Name: "Ann", LastSignInAt: at(time.Hour)},
		DiscordLinkedAt:       at(time.Minute),
		Games:                 []GameRow{{ID: g, Name: "T", State: "ended", CreatedAt: t0, TheirSeat: &seat}},
		GamesTruncated:        true,
		Decks:                 []DeckRow{{ID: "d", Name: "Deck", Format: "text"}},
		DeckRequests:          []DeckRequestRow{{DeckKey: "moxfield:x", AskedAt: at(5 * time.Minute)}},
		DeckRequestsTruncated: true,
	}
	got := MergeAccount(rows, []string{"u1"}, Overlay{}, t0)
	if !got.Account.PlayingNow || got.SignIn.RevokePath != "/admin/users/u1/revoke-sessions" ||
		got.SignIn.LastSignInAt != at(time.Hour).UnixMilli() || got.SignIn.DiscordLinkedAt == 0 || got.SignIn.SessionsInvalidBefore != 0 {
		t.Errorf("account / sign-in = %+v / %+v", got.Account, got.SignIn)
	}
	if len(got.Games) != 1 || got.Games[0].TheirSeat == nil || *got.Games[0].TheirSeat != 1 || !got.GamesTruncated {
		t.Errorf("games = %+v", got.Games)
	}
	if len(got.Decks) != 1 || got.Decks[0].Commanders == nil {
		t.Errorf("decks = %+v; commanders must be [] when none", got.Decks)
	}
	if len(got.DeckRequests) != 1 || !got.DeckRequestsTruncated {
		t.Errorf("deck requests = %+v", got.DeckRequests)
	}
}

// A loaded table's row takes memory's state, kinds, agent client and
// host on top of its row; an unloaded one is the row alone.
func TestMergeGameOverlaysMemory(t *testing.T) {
	id := uuid.New()
	alice, agent, late := uuid.New(), uuid.New(), uuid.New()
	row := GameRow{
		ID: id, Name: "T", State: "lobby", CreatedAt: t0, ArchivedAt: at(time.Hour), HostPlayerID: agent.String(),
		Seats: []SeatRow{
			{Seat: 0, PlayerID: alice.String(), GuestName: "Alice"},
			// A row from before migration 0010: no agent_client.
			{Seat: 1, PlayerID: agent.String(), GuestName: "Claude"},
		},
	}
	live := LiveTable{ID: id, State: "active", Archived: false, Seats: []LiveSeat{
		{Seat: 0, PlayerID: alice, Kind: KindHuman, Name: "Alice", Host: true},
		{Seat: 1, PlayerID: agent, Kind: KindAgent, Name: "Claude", AgentClient: "codex"},
		{Seat: 2, PlayerID: late, Kind: KindHuman, Name: "Late", DeckName: "Fresh"},
	}}

	plain := MergeGame(row, Overlay{})
	if plain.Loaded || plain.State != "lobby" || plain.Seats[1].Kind != KindHuman || !plain.Seats[1].Host || plain.ArchivedAt == 0 {
		t.Errorf("unloaded = %+v", plain)
	}

	got := MergeGame(row, Overlay{Tables: []LiveTable{{ID: uuid.New()}, live}})
	if !got.Loaded || got.State != "active" || got.ArchivedAt != 0 {
		t.Errorf("loaded: loaded %v state %q archived %d; want memory's", got.Loaded, got.State, got.ArchivedAt)
	}
	if len(got.Seats) != 3 {
		t.Fatalf("%d seats, want the row's two and memory's newer one", len(got.Seats))
	}
	if s := got.Seats[0]; !s.Host || s.Kind != KindHuman {
		t.Errorf("seat 0 = %+v, want memory's host", s)
	}
	if s := got.Seats[1]; s.Kind != KindAgent || s.AgentClient != "codex" || s.Host {
		t.Errorf("seat 1 = %+v, want the agent badge from memory", s)
	}
	if s := got.Seats[2]; s.GuestName != "Late" || s.DeckName != "Fresh" {
		t.Errorf("seat 2 = %+v", s)
	}
	if got.SpectatorsConnected != nil || got.Seats[0].Connected != nil {
		t.Error("connected counts served without the sockets")
	}
}

func TestConnectedCountsWhenTheSocketsAreKnown(t *testing.T) {
	id, other := uuid.New(), uuid.New()
	p0, p1 := uuid.New(), uuid.New()
	live := LiveTable{ID: id, State: "active", Seats: []LiveSeat{
		{Seat: 0, PlayerID: p0, Kind: KindHuman, Name: "A"},
		{Seat: 1, PlayerID: p1, Kind: KindBot, Name: "Bot", BotTier: "random"},
	}}
	ov := Overlay{Tables: []LiveTable{live}, SocketsKnown: true, Sockets: []Socket{
		{GameID: id, PlayerID: p0},
		{GameID: id, PlayerID: p0},
		{GameID: id, ReadOnly: true},
		{GameID: id, ReadOnly: true, Admin: true}, // an admin watching, not a spectator
		{GameID: other, ReadOnly: true},
	}}
	got := LiveGame(live, ov)
	if got.SpectatorsConnected == nil || *got.SpectatorsConnected != 1 {
		t.Errorf("spectators = %v, want 1", got.SpectatorsConnected)
	}
	if c := got.Seats[0].Connected; c == nil || *c != 2 {
		t.Errorf("seat 0 connected = %v, want 2", c)
	}
	if c := got.Seats[1].Connected; c == nil || *c != 0 {
		t.Errorf("bot seat connected = %v, want a known 0", c)
	}
}

func TestLiveGameNamesItsAccounts(t *testing.T) {
	id, p := uuid.New(), uuid.New()
	live := LiveTable{ID: id, Name: "Practice", State: "active", Practice: true, CreatedAt: t0, Seats: []LiveSeat{
		{Seat: 1, PlayerID: uuid.New(), Kind: KindBot, Name: "Practice Bot", BotTier: "random"},
		{Seat: 0, PlayerID: p, Kind: KindHuman, UserID: "u1", Name: "alice", DisplayName: "Alice"},
	}}
	got := LiveGame(live, Overlay{Accounts: map[string]AccountRef{"u1": {ID: "u1", Name: "Alice A.", AvatarURL: "/avatars/1/a.png"}}})
	if !got.Practice || !got.Loaded || got.CreatedAt != t0.UnixMilli() || got.Seats[0].Seat != 0 {
		t.Errorf("practice game = %+v", got)
	}
	if a := got.Seats[0].Account; a == nil || a.Name != "Alice A." || a.AvatarURL == "" {
		t.Errorf("account = %+v, want the users row's", a)
	}
	if s := got.Seats[1]; s.GuestName != "Practice Bot" || s.Kind != KindBot || s.BotTier != "random" {
		t.Errorf("bot seat = %+v", s)
	}
	// With no users row, the seat's own name.
	got = LiveGame(live, Overlay{})
	if a := got.Seats[0].Account; a == nil || a.Name != "Alice" || a.ID != "u1" {
		t.Errorf("account without a row = %+v", a)
	}
}

func TestMergeGamesPractice(t *testing.T) {
	row := GameRow{ID: uuid.New(), Name: "Row", State: "active", CreatedAt: t0}
	older := LiveTable{ID: uuid.New(), Name: "Older", State: "active", Practice: true, CreatedAt: at(1 * time.Minute)}
	newer := LiveTable{ID: uuid.New(), Name: "Newer", State: "active", Practice: true, CreatedAt: at(2 * time.Minute)}
	loaded := LiveTable{ID: row.ID, State: "active"}
	ov := Overlay{Tables: []LiveTable{older, loaded, newer}}
	next := &Cursor{CreatedAt: t0, ID: row.ID}
	res := GamesResult{Rows: []GameRow{row}, Next: next}

	names := func(r GamesResponse) []string {
		var out []string
		for _, g := range r.Games {
			out = append(out, g.Name)
		}
		return out
	}
	cases := []struct {
		name      string
		f         GamesFilter
		firstPage bool
		want      []string
		next      bool
	}{
		{"exclude", GamesFilter{Practice: PracticeExclude}, true, []string{"Row"}, true},
		{"include, first page", GamesFilter{Practice: PracticeInclude}, true, []string{"Newer", "Older", "Row"}, true},
		{"include, later page", GamesFilter{Practice: PracticeInclude}, false, []string{"Row"}, true},
		{"only", GamesFilter{Practice: PracticeOnly}, true, []string{"Newer", "Older"}, false},
		{"only, filtered out", GamesFilter{Practice: PracticeOnly, State: "ended"}, true, nil, false},
	}
	for _, tc := range cases {
		got := MergeGames(res, ov, tc.f, tc.firstPage, t0)
		if !slices.Equal(names(got), tc.want) || (got.NextCursor != "") != tc.next {
			t.Errorf("%s: %v next %q, want %v next %v", tc.name, names(got), got.NextCursor, tc.want, tc.next)
		}
		if got.Games == nil {
			t.Errorf("%s: games is null, want []", tc.name)
		}
	}
}

func TestMatchesLive(t *testing.T) {
	yes, no := true, false
	tbl := LiveTable{State: "active", Archived: true, Seats: []LiveSeat{{UserID: "u1"}}}
	cases := []struct {
		f    GamesFilter
		want bool
	}{
		{GamesFilter{}, true},
		{GamesFilter{State: "active"}, true},
		{GamesFilter{State: "lobby"}, false},
		{GamesFilter{Archived: &yes}, true},
		{GamesFilter{Archived: &no}, false},
		{GamesFilter{UserID: "u1"}, true},
		{GamesFilter{UserID: "u2"}, false},
	}
	for _, tc := range cases {
		if got := tc.f.MatchesLive(tbl); got != tc.want {
			t.Errorf("%+v: %v, want %v", tc.f, got, tc.want)
		}
	}
}

func TestParseGamesFilter(t *testing.T) {
	get := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	f, err := ParseGamesFilter(get(nil))
	if err != nil || f.Practice != PracticeExclude || f.State != "" || f.Archived != nil || f.UserID != "" {
		t.Errorf("no parameters: %+v %v; want any, with practice excluded", f, err)
	}
	user := uuid.New()
	f, err = ParseGamesFilter(get(map[string]string{"state": "ended", "archived": "false", "practice": "only", "user": user.String()}))
	if err != nil || f.State != "ended" || f.Archived == nil || *f.Archived || f.Practice != PracticeOnly || f.UserID != user.String() {
		t.Errorf("every filter: %+v %v", f, err)
	}
	for param, bad := range map[string]string{"state": "running", "archived": "1", "practice": "yes", "user": "bob"} {
		_, err := ParseGamesFilter(get(map[string]string{param: bad}))
		if err == nil || !strings.HasPrefix(err.Error(), param+" ") {
			t.Errorf("%s=%s: %v, want an error naming %s", param, bad, err, param)
		}
	}
}

func TestParseGamesLimit(t *testing.T) {
	for in, want := range map[string]int{"": DefaultGamesLimit, "1": 1, "200": 200, "5000": MaxGamesLimit} {
		if got, err := ParseGamesLimit(in); err != nil || got != want {
			t.Errorf("limit %q: %d %v, want %d", in, got, err, want)
		}
	}
	for _, bad := range []string{"0", "-3", "ten", "1.5"} {
		if _, err := ParseGamesLimit(bad); err == nil || !strings.HasPrefix(err.Error(), "limit ") {
			t.Errorf("limit %q: %v, want an error naming limit", bad, err)
		}
	}
}

func TestCursorRoundTrip(t *testing.T) {
	c := Cursor{CreatedAt: time.UnixMilli(1759600000123), ID: uuid.New()}
	got, err := ParseCursor(c.String())
	if err != nil || !got.CreatedAt.Equal(c.CreatedAt) || got.ID != c.ID {
		t.Errorf("round trip: %+v %v, want %+v", got, err, c)
	}
	for _, bad := range []string{"", "!!", "bm90LWEtY3Vyc29y", c.String() + "x"} {
		if _, err := ParseCursor(bad); err == nil {
			t.Errorf("cursor %q parsed", bad)
		}
	}
}
