package lobby

// admin_views_test.go covers ADR 0124 §8 for the admin views of
// accounts and games: the field allowlist over a real database, the
// filters' 400s, the 503 with no database, the 404s, and the audit
// line that never carries a query string. The gate itself (401, 403,
// player mode) is in admins_test.go and player_mode_test.go.

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/adminview"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decklibrary"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/playmat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// emptyAdminViews is a store with no rows: enough for the gate's probes
// to reach a handler and get its answer.
type emptyAdminViews struct{}

func (emptyAdminViews) Accounts(context.Context, adminview.AccountsQuery) (adminview.AccountsResult, error) {
	return adminview.AccountsResult{}, nil
}

func (emptyAdminViews) Account(context.Context, uuid.UUID) (adminview.AccountRows, error) {
	return adminview.AccountRows{}, adminview.ErrNotFound
}

func (emptyAdminViews) Games(context.Context, adminview.GamesQuery) (adminview.GamesResult, error) {
	return adminview.GamesResult{}, nil
}

func (emptyAdminViews) Game(context.Context, uuid.UUID) (adminview.GameRow, error) {
	return adminview.GameRow{}, adminview.ErrNotFound
}

func (emptyAdminViews) AccountRefs(context.Context, []string) (map[string]adminview.AccountRef, error) {
	return map[string]adminview.AccountRef{}, nil
}

// fixedLiveSockets is a hub with these connections.
type fixedLiveSockets []ws.LiveSocket

func (f fixedLiveSockets) LiveSockets() []ws.LiveSocket { return f }

// --- the seeded server -------------------------------------------------

const (
	annSnowflake     = "424242424242424242"
	pendingSnowflake = "333333333333333333"
	refreshMarker    = "REFRESH-TOKEN-MARKER"
	deckListMarker   = "1 Secret Listed Card"
)

// adminViewsWorld is a durable server (a database and a lobby over it)
// seeded with everything ADR 0124 §4 says the views must never serve.
type adminViewsWorld struct {
	s        *adminStack
	d        *db.DB
	l        *Lobby
	token    string
	ann      uuid.UUID
	live     GameMeta
	ended    uuid.UUID
	practice GameMeta
	secrets  []string // values that may appear nowhere in any answer
}

func newAdminViewsWorld(t *testing.T) *adminViewsWorld {
	t.Helper()
	dir := t.TempDir()
	d := openTestDB(t, dir)
	l := NewLobbyWithStore(ws.NewRoomManager(quietLogger(), dir), NewSQLStore(d))
	l.SetBotHost(newFakeBotHost())
	w := &adminViewsWorld{d: d, l: l, ann: uuid.New()}
	ctx := context.Background()
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := d.Exec(q, args...); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	now := time.Now().UnixMilli()

	// Ann: an account with a Discord identity, a sealed refresh token
	// and scopes, a saved deck and a deck request.
	exec(`INSERT INTO users (id, display_name, avatar_url, created_at, last_seen_at, sessions_invalid_before) VALUES (?, 'Ann', ?, ?, ?, ?)`,
		w.ann.String(), "/avatars/"+annSnowflake+"/abcdef.png", now-1000, now-500, now-100)
	exec(`INSERT INTO identities (provider, subject, user_id, display_name, refresh_token, scopes, linked_at)
	      VALUES ('discord', ?, ?, 'Ann', ?, 'identify', ?)`, annSnowflake, w.ann.String(), []byte(refreshMarker), now-1000)
	if _, err := decklibrary.NewSQLStore(d).UpsertFromLink(ctx, w.ann, "Ann's deck", "moxfield", "Commander:\n1 Atraxa\n"+deckListMarker,
		"https://moxfield.com/decks/abc", []string{"Atraxa"}, 100); err != nil {
		t.Fatal(err)
	}
	exec(`INSERT INTO deck_requests (deck_key, issue_number, issue_url, created_at) VALUES ('moxfield:abc', 12, 'https://github.com/o/r/issues/12', ?)`, now)
	exec(`INSERT INTO deck_request_asks (deck_key, requester, at) VALUES ('moxfield:abc', ?, ?)`, "discord:"+annSnowflake, now)

	// A loaded table Ann created and sits at, with a guest, a Discord
	// seat waiting for its person, an agent and a bot.
	var err error
	if w.live, err = l.CreateBy("Live table", w.ann); err != nil {
		t.Fatal(err)
	}
	_, annSeat, err := l.JoinAs(w.live.ID, w.live.InviteToken, "Ann", DiscordIdentity{ID: annSnowflake, Username: "ann", GlobalName: "Ann", AvatarHash: "abcdef"}, w.ann)
	if err != nil {
		t.Fatal(err)
	}
	_, gus, err := l.Join(w.live.ID, w.live.InviteToken, "Gus")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = l.JoinWithIdentity(w.live.ID, w.live.InviteToken, "Dee", DiscordIdentity{ID: pendingSnowflake, Username: "dee"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err = l.JoinAgent(w.live.ID, w.live.InviteToken, "Claude", AgentDecl{Client: "Codex"}); err != nil {
		t.Fatal(err)
	}
	ticket, err := l.MintReclaim(w.live.ID, gus)
	if err != nil {
		t.Fatal(err)
	}

	// An ended table only the database holds, whose agent seat is known
	// from migration 0010's column alone.
	w.ended = uuid.New()
	exec(`INSERT INTO games (id, name, created_by, state, created_at, started_at, ended_at, winner_seat, outcome, host_player_id)
	      VALUES (?, 'Ended table', ?, 'ended', ?, ?, ?, 0, 'win', ?)`, w.ended.String(), w.ann.String(), now-5000, now-4000, now-3000, "p-ann")
	exec(`INSERT INTO seats (game_id, seat, player_id, user_id, guest_name, deck_name) VALUES (?, 0, 'p-ann', ?, 'Ann', 'Ann''s deck')`, w.ended.String(), w.ann.String())
	exec(`INSERT INTO seats (game_id, seat, player_id, guest_name, agent_client) VALUES (?, 1, 'p-agent', 'Claude', 'claude-code')`, w.ended.String())
	exec(`INSERT INTO seats (game_id, seat, player_id, guest_name, bot_tier) VALUES (?, 2, 'p-bot', 'Bot', 'heuristic')`, w.ended.String())
	exec(`INSERT INTO invites (token_hash, game_id, kind, created_at) VALUES (X'DEADBEEF', ?, 'player', ?)`, w.ended.String(), now)

	// Ann's practice table: memory only.
	human, bot := practiceSeats("user:" + w.ann.String())
	human.UserID = w.ann
	if w.practice, _, err = l.CreatePractice(human, bot); err != nil {
		t.Fatal(err)
	}

	// Connected at the live table: Ann at her seat, a guest spectator,
	// Ann watching from a second tab, and the shared token. Plus one
	// socket at the practice table.
	connected := time.Now().Add(-time.Minute)
	sockets := fixedLiveSockets{
		{GameID: w.live.ID, PlayerID: annSeat, UserID: w.ann, ConnectedAt: connected},
		{GameID: w.live.ID, ReadOnly: true, ConnectedAt: connected.Add(time.Second)},
		{GameID: w.live.ID, UserID: w.ann, ReadOnly: true, ConnectedAt: connected.Add(2 * time.Second)},
		{GameID: w.live.ID, Admin: true, ConnectedAt: connected.Add(3 * time.Second)},
		{GameID: w.practice.ID, ReadOnly: true, ConnectedAt: connected},
	}
	svc := playmat.NewService(playmat.NewFileStore(filepath.Join(dir, "playmats")), d.DB, nil)
	w.s = newAdminStackIn(t, "", nil, func(c *Config) {
		c.Lobby = l
		c.AdminViews = adminview.NewSQLStore(d)
		c.LiveSockets = sockets
		c.Playmats = svc
	})
	w.token = w.s.adminToken(t)
	if _, err := svc.SetFromBytes(ctx, w.ann, pngOf(t, 64, 48)); err != nil {
		t.Fatal(err)
	}
	w.secrets = []string{
		w.live.InviteToken, w.live.SpectatorInvite, ticket.Token, deckListMarker, refreshMarker,
		"discord:" + annSnowflake, pendingSnowflake, "DEADBEEF", "deadbeef",
	}
	return w
}

func (w *adminViewsWorld) get(t *testing.T, path string) (int, []byte) {
	t.Helper()
	resp := do(t, w.s.srv, "GET", path, w.token, nil)
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, raw
}

// --- the field allowlist -------------------------------------------------

// gameRowFields are a table row's paths (§3.3), under prefix.
func gameRowFields(prefix string) []string {
	out := []string{}
	for _, f := range []string{
		"id", "name", "state", "created_at", "started_at", "ended_at", "archived_at", "outcome", "winner_seat",
		"creator.id", "creator.name", "practice", "loaded", "spectators_connected", "seats",
		"seats[].seat", "seats[].kind", "seats[].account.id", "seats[].account.name", "seats[].account.avatar_url",
		"seats[].guest_name", "seats[].discord_pending", "seats[].bot_tier", "seats[].agent_client",
		"seats[].deck_name", "seats[].host", "seats[].connected",
	} {
		out = append(out, prefix+f)
	}
	return out
}

func accountFields(prefix string) []string {
	out := []string{}
	for _, f := range []string{"id", "name", "avatar_url", "first_seen_at", "last_sign_in_at", "games_played", "last_played_at", "playing_now"} {
		out = append(out, prefix+f)
	}
	return out
}

func fieldSet(groups ...[]string) map[string]bool {
	out := map[string]bool{}
	for _, g := range groups {
		for _, f := range g {
			out[f] = true
		}
	}
	return out
}

// jsonLeaves lists every leaf path of v ("a.b", "a[].c", an empty array
// as its own path) and every string value with the key it sits under.
func jsonLeaves(v any, prefix, key string, paths map[string]bool, strs map[string][]string) {
	switch x := v.(type) {
	case map[string]any:
		for k, vv := range x {
			p := k
			if prefix != "" {
				p = prefix + "." + k
			}
			jsonLeaves(vv, p, k, paths, strs)
		}
	case []any:
		if len(x) == 0 {
			paths[prefix] = true
		}
		for _, vv := range x {
			jsonLeaves(vv, prefix+"[]", key, paths, strs)
		}
	default:
		paths[prefix] = true
		if s, ok := x.(string); ok {
			strs[key] = append(strs[key], s)
		}
	}
}

// TestAdminViewsServeOnlyTheirFields is ADR 0124 §8's field allowlist:
// each route, asked by an admin over a database seeded with everything
// §4 says is never shown, serves exactly its pinned fields, no key that
// names a secret, the seeded snowflake only inside an avatar path, and
// none of the seeded secrets at all.
func TestAdminViewsServeOnlyTheirFields(t *testing.T) {
	w := newAdminViewsWorld(t)
	gameDetail := fieldSet([]string{"generated_at", "connections", "connections[].kind", "connections[].seat",
		"connections[].account.id", "connections[].account.name", "connections[].account.avatar_url", "connections[].since"},
		gameRowFields(""))
	routes := []struct {
		path    string
		allowed map[string]bool
		// must is a sample of paths the seed has to produce, so the
		// allowlist is not passing over an empty answer.
		must []string
	}{
		{"/admin/users", fieldSet([]string{"generated_at", "truncated"}, accountFields("accounts[].")),
			[]string{"accounts[].avatar_url", "accounts[].playing_now", "accounts[].games_played"}},
		{"/admin/users/" + w.ann.String(), fieldSet(
			[]string{"generated_at", "games_truncated", "deck_requests_truncated",
				"playmat_url", "sign_in.last_sign_in_at", "sign_in.discord_linked_at", "sign_in.sessions_invalid_before", "sign_in.revoke_path",
				"decks", "decks[].id", "decks[].name", "decks[].format", "decks[].source_url", "decks[].commanders", "decks[].commanders[]",
				"decks[].card_count", "decks[].created_at", "decks[].updated_at",
				"deck_requests", "deck_requests[].deck_key", "deck_requests[].asked_at", "deck_requests[].issue_number", "deck_requests[].issue_url",
				"games", "games[].their_seat"},
			accountFields("account."), gameRowFields("games[].")),
			[]string{"playmat_url", "sign_in.sessions_invalid_before", "decks[].commanders[]", "deck_requests[].issue_url", "games[].their_seat",
				"games[].seats[].agent_client", "games[].seats[].discord_pending", "games[].creator.name"}},
		{"/admin/games?practice=include", fieldSet([]string{"generated_at", "next_cursor", "games"}, gameRowFields("games[].")),
			[]string{"games[].practice", "games[].seats[].agent_client", "games[].seats[].bot_tier", "games[].seats[].account.avatar_url",
				"games[].outcome", "games[].winner_seat"}},
		{"/admin/games/" + w.live.ID.String(), gameDetail, []string{"seats[].discord_pending", "seats[].agent_client", "creator.id", "loaded",
			"seats[].connected", "spectators_connected", "connections[].kind", "connections[].seat", "connections[].account.avatar_url", "connections[].since"}},
		{"/admin/games/" + w.ended.String(), gameDetail, []string{"seats[].agent_client", "outcome", "ended_at"}},
		{"/admin/games/" + w.practice.ID.String(), gameDetail, []string{"practice", "seats[].account.name", "seats[].bot_tier"}},
	}
	for _, rt := range routes {
		code, raw := w.get(t, rt.path)
		if code != http.StatusOK {
			t.Errorf("GET %s: %d %s", rt.path, code, raw)
			continue
		}
		for _, secret := range w.secrets {
			if secret != "" && strings.Contains(string(raw), secret) {
				t.Errorf("GET %s serves %q, which §4 never shows", rt.path, secret)
			}
		}
		var body any
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("GET %s: %v", rt.path, err)
		}
		paths, strs := map[string]bool{}, map[string][]string{}
		jsonLeaves(body, "", "", paths, strs)
		var extra []string
		for p := range paths {
			if !rt.allowed[p] {
				extra = append(extra, p)
			}
			for _, key := range strings.FieldsFunc(p, func(r rune) bool { return r == '.' || r == '[' || r == ']' }) {
				for _, bad := range forbiddenKeyParts {
					if strings.Contains(key, bad) {
						t.Errorf("GET %s serves the key %q (in %s), which names %q", rt.path, key, p, bad)
					}
				}
			}
		}
		sort.Strings(extra)
		if len(extra) > 0 {
			t.Errorf("GET %s serves fields not on its pinned list (ADR 0124 §3; add them deliberately): %v", rt.path, extra)
		}
		for _, p := range rt.must {
			if !paths[p] {
				t.Errorf("GET %s does not serve %s; the seed is not reaching it", rt.path, p)
			}
		}
		for key, values := range strs {
			for _, v := range values {
				if strings.Contains(v, annSnowflake) && key != "avatar_url" {
					t.Errorf("GET %s serves the snowflake under %q: %q", rt.path, key, v)
				}
			}
		}
	}
}

// --- what the routes answer ----------------------------------------------

func TestAdminViewsAnswerTheirRows(t *testing.T) {
	w := newAdminViewsWorld(t)

	var games adminview.GamesResponse
	code, raw := w.get(t, "/admin/games")
	if code != http.StatusOK || json.Unmarshal(raw, &games) != nil {
		t.Fatalf("GET /admin/games: %d %s", code, raw)
	}
	if len(games.Games) != 2 || games.Games[0].ID != w.live.ID.String() || games.Games[1].ID != w.ended.String() {
		t.Fatalf("games = %s; want the live table then the ended one, no practice table", raw)
	}
	live := games.Games[0]
	if !live.Loaded || live.Practice || live.Creator == nil || live.Creator.Name != "Ann" || len(live.Seats) != 4 {
		t.Errorf("live table = %+v", live)
	}
	kinds := map[string]bool{}
	for _, s := range live.Seats {
		kinds[s.Kind] = true
		if s.Kind == adminview.KindAgent && s.AgentClient != "codex" {
			t.Errorf("agent seat = %+v", s)
		}
		if s.GuestName == "Dee" && !s.DiscordPending {
			t.Errorf("Dee's seat is not pending: %+v", s)
		}
	}
	if !kinds[adminview.KindAgent] || !kinds[adminview.KindHuman] {
		t.Errorf("kinds = %v", kinds)
	}
	if ended := games.Games[1]; ended.Loaded || ended.Outcome != adminview.OutcomeWin || ended.Seats[1].Kind != adminview.KindAgent || ended.Seats[2].BotTier != "heuristic" {
		t.Errorf("ended table = %+v", ended)
	}

	code, raw = w.get(t, "/admin/games?practice=only")
	if code != http.StatusOK || json.Unmarshal(raw, &games) != nil || len(games.Games) != 1 || games.Games[0].ID != w.practice.ID.String() {
		t.Errorf("practice=only: %d %s", code, raw)
	}
	code, raw = w.get(t, "/admin/games?state=ended&user="+w.ann.String())
	if code != http.StatusOK || json.Unmarshal(raw, &games) != nil || len(games.Games) != 1 || games.Games[0].ID != w.ended.String() {
		t.Errorf("state=ended&user=Ann: %d %s", code, raw)
	}

	// Pages: one table a page, then the next with the cursor.
	code, raw = w.get(t, "/admin/games?limit=1")
	if code != http.StatusOK || json.Unmarshal(raw, &games) != nil || len(games.Games) != 1 || games.NextCursor == "" {
		t.Fatalf("limit=1: %d %s", code, raw)
	}
	cursor := games.NextCursor
	games = adminview.GamesResponse{}
	code, raw = w.get(t, "/admin/games?limit=1&cursor="+cursor)
	if code != http.StatusOK || json.Unmarshal(raw, &games) != nil || len(games.Games) != 1 || games.Games[0].ID != w.ended.String() || games.NextCursor != "" {
		t.Errorf("the second page: %d %s", code, raw)
	}

	var accounts adminview.AccountsResponse
	code, raw = w.get(t, "/admin/users?played=7d")
	if code != http.StatusOK || json.Unmarshal(raw, &accounts) != nil || len(accounts.Accounts) != 1 {
		t.Fatalf("played=7d: %d %s", code, raw)
	}
	if a := accounts.Accounts[0]; a.ID != w.ann.String() || a.GamesPlayed != 1 || a.LastPlayedAt == 0 || a.PlayingNow {
		t.Errorf("Ann = %+v; her ended table counts, and the live table has not started", a)
	}

	var account adminview.AccountResponse
	code, raw = w.get(t, "/admin/users/"+w.ann.String())
	if code != http.StatusOK || json.Unmarshal(raw, &account) != nil {
		t.Fatalf("Ann's account: %d %s", code, raw)
	}
	if len(account.Games) != 2 || len(account.Decks) != 1 || len(account.DeckRequests) != 1 || account.DeckRequests[0].IssueNumber != 12 ||
		account.SignIn.RevokePath != "/admin/users/"+w.ann.String()+"/revoke-sessions" || account.SignIn.SessionsInvalidBefore == 0 {
		t.Errorf("Ann's account = %s", raw)
	}
}

// The table detail lists the table's live connections, earliest first,
// and every row of a loaded table counts them by GET /admin/live's
// rules: a seat's sockets, and read-only sockets that are not admins'.
func TestAdminViewsCountTheLiveConnections(t *testing.T) {
	w := newAdminViewsWorld(t)
	var detail adminview.GameResponse
	code, raw := w.get(t, "/admin/games/"+w.live.ID.String())
	if code != http.StatusOK || json.Unmarshal(raw, &detail) != nil || detail.Connections == nil {
		t.Fatalf("GET the live table: %d %s", code, raw)
	}
	conns := *detail.Connections
	type want struct {
		kind    string
		seated  bool
		account bool
	}
	wants := []want{{adminview.ConnSeat, true, true}, {adminview.ConnSpectator, false, false}, {adminview.ConnSpectator, false, true}, {adminview.ConnAdmin, false, false}}
	if len(conns) != len(wants) {
		t.Fatalf("connections = %s, want %d", raw, len(wants))
	}
	for i, wnt := range wants {
		c := conns[i]
		if c.Kind != wnt.kind || (c.Seat != nil) != wnt.seated || (c.Account != nil) != wnt.account || c.Since == 0 {
			t.Errorf("connection %d = %+v (seat %v, account %+v), want %+v", i, c, c.Seat, c.Account, wnt)
		}
		if c.Account != nil && (c.Account.ID != w.ann.String() || c.Account.Name != "Ann" || c.Account.AvatarURL == "") {
			t.Errorf("connection %d's account = %+v, want Ann's row", i, c.Account)
		}
	}
	if detail.SpectatorsConnected == nil || *detail.SpectatorsConnected != 2 {
		t.Errorf("spectators_connected = %v, want 2", detail.SpectatorsConnected)
	}
	for _, s := range detail.Seats {
		want := 0
		if s.Account != nil && s.Account.ID == w.ann.String() {
			want = 1
		}
		if s.Connected == nil || *s.Connected != want {
			t.Errorf("seat %d connected = %v, want %d", s.Seat, s.Connected, want)
		}
	}

	// A table only the database holds has no connections, and says so.
	code, raw = w.get(t, "/admin/games/"+w.ended.String())
	detail = adminview.GameResponse{}
	if code != http.StatusOK || json.Unmarshal(raw, &detail) != nil || detail.Connections == nil || len(*detail.Connections) != 0 || detail.SpectatorsConnected != nil {
		t.Errorf("the ended table: %d %s; want connections [] and no counts", code, raw)
	}
	// The practice table, from memory: its one spectator.
	code, raw = w.get(t, "/admin/games/"+w.practice.ID.String())
	detail = adminview.GameResponse{}
	if code != http.StatusOK || json.Unmarshal(raw, &detail) != nil || detail.SpectatorsConnected == nil || *detail.SpectatorsConnected != 1 {
		t.Errorf("the practice table: %d %s", code, raw)
	}
}

func TestAdminViewsRefuseBadRequests(t *testing.T) {
	w := newAdminViewsWorld(t)
	cases := []struct {
		path string
		want int
		msg  string
	}{
		{"/admin/users?played=2d", http.StatusBadRequest, "played"},
		{"/admin/games?state=running", http.StatusBadRequest, "state"},
		{"/admin/games?archived=yes", http.StatusBadRequest, "archived"},
		{"/admin/games?practice=all", http.StatusBadRequest, "practice"},
		{"/admin/games?user=bob", http.StatusBadRequest, "user"},
		{"/admin/games?limit=0", http.StatusBadRequest, "limit"},
		{"/admin/games?cursor=nope", http.StatusBadRequest, "cursor"},
		{"/admin/users/not-a-uuid", http.StatusBadRequest, "user id"},
		{"/admin/users/" + uuid.NewString(), http.StatusNotFound, "account not found"},
		{"/admin/games/not-a-uuid", http.StatusBadRequest, "game id"},
		{"/admin/games/" + uuid.NewString(), http.StatusNotFound, "game not found"},
	}
	for _, tc := range cases {
		code, raw := w.get(t, tc.path)
		var body struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(raw, &body)
		if code != tc.want || !strings.Contains(body.Error, tc.msg) {
			t.Errorf("GET %s: %d %q, want %d naming %q", tc.path, code, body.Error, tc.want, tc.msg)
		}
	}
	// A missing or empty parameter means any.
	for _, path := range []string{"/admin/users?played=", "/admin/games?state=&archived=&practice=&user=&limit=&cursor="} {
		if code, raw := w.get(t, path); code != http.StatusOK {
			t.Errorf("GET %s: %d %s", path, code, raw)
		}
	}
}

// With no database there is no store, and the four routes say so; the
// gate still comes first.
func TestAdminViewsNeedTheDatabase(t *testing.T) {
	s := newAdminStack(t)
	token := s.adminToken(t)
	for _, path := range []string{"/admin/users", "/admin/users/" + uuid.NewString(), "/admin/games", "/admin/games/" + uuid.NewString()} {
		resp := do(t, s.srv, "GET", path, token, nil)
		var body struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusServiceUnavailable || body.Error != "admin views need the user database, and this server has none" {
			t.Errorf("GET %s with no database: %d %q", path, resp.StatusCode, body.Error)
		}
		if got := status(t, s.srv, "GET", path, "", nil); got != http.StatusUnauthorized {
			t.Errorf("GET %s with no session: %d, want 401", path, got)
		}
	}
}

// The audit line names the path and never the query string, so a
// filter is never logged (ADR 0017 §9, ADR 0124 §4).
func TestAdminViewsLogNoQueryString(t *testing.T) {
	s := newAdminStackWith(t, nil, func(c *Config) { c.AdminViews = emptyAdminViews{} })
	token := s.adminToken(t)
	if got := status(t, s.srv, "GET", "/admin/games?state=active&user="+uuid.NewString(), token, nil); got != http.StatusOK {
		t.Fatalf("GET /admin/games: %d", got)
	}
	var found bool
	for _, line := range strings.Split(s.log.String(), "\n") {
		if !strings.Contains(line, "admin action") || !strings.Contains(line, "/admin/games") {
			continue
		}
		found = true
		if strings.Contains(line, "?") || strings.Contains(line, "state=") {
			t.Errorf("the audit line carries the query string: %s", line)
		}
	}
	if !found {
		t.Errorf("no admin action line for GET /admin/games:\n%s", s.log.String())
	}
}

// The account view names the playmat the admin could remove, and stops
// naming it once DELETE /admin/users/{id}/playmat has run.
func TestAdminAccountViewShowsPlaymatUntilRemoved(t *testing.T) {
	w := newAdminViewsWorld(t)
	path := "/admin/users/" + w.ann.String()
	read := func() string {
		code, raw := w.get(t, path)
		if code != http.StatusOK {
			t.Fatalf("GET %s: %d %s", path, code, raw)
		}
		var body struct {
			PlaymatURL string `json:"playmat_url"`
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatal(err)
		}
		return body.PlaymatURL
	}
	if u := read(); !strings.HasPrefix(u, "/playmats/") {
		t.Fatalf("playmat_url = %q, want a /playmats/ path", u)
	}
	resp := do(t, w.s.srv, "DELETE", path+"/playmat", w.token, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE playmat = %d", resp.StatusCode)
	}
	if u := read(); u != "" {
		t.Errorf("playmat_url = %q after removal, want none", u)
	}
}
