package lobby

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// GET /admin/live, the admin views' Live now (ADR 0124 §3.4, §8): the
// fields it serves, the totals it shares with the Overview's tiles,
// and the locks it takes. The answer-table probe and the admin-route
// table (player_mode_test.go, admins_test.go) cover who may call it;
// TestScrapeDuringTableTraffic runs it under -race beside table
// traffic.

// liveNowStack is the HTTP and WebSocket stack with the tables
// collector over the same lobby and hub, as main wires them.
type liveNowStack struct {
	srv   *httptest.Server
	lobby *Lobby
	hub   *ws.Hub
	auth  auth.Authenticator
	reg   *prometheus.Registry
}

func newLiveNowStack(t *testing.T, configure func(*Config)) *liveNowStack {
	t.Helper()
	quiet := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := auth.NewMemoryAuthenticator()
	mgr := ws.NewRoomManager(quiet, "")
	l := NewLobby(mgr)
	l.SetBotHost(newFakeBotHost())
	hub := ws.NewHub(quiet)
	hub.SetManager(mgr)
	hub.SetAuthorizer(&WSAuthorizer{Auth: a})
	l.SetStateBroadcaster(hub)
	cfg := Config{Lobby: l, Auth: a, AdminToken: "shared-admin-token", LiveSockets: hub, Log: quiet}
	if configure != nil {
		configure(&cfg)
	}
	mux := http.NewServeMux()
	mux.Handle("/", Handler(cfg))
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	reg := prometheus.NewRegistry()
	reg.MustRegister(metrics.NewTablesCollector(l, hub))
	return &liveNowStack{srv: srv, lobby: l, hub: hub, auth: a, reg: reg}
}

func (s *liveNowStack) issue(t *testing.T, p auth.Principal) string {
	t.Helper()
	tok, _, err := s.auth.Issue(context.Background(), p, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

// seatToken is a guest seat's session.
func (s *liveNowStack) seatToken(t *testing.T, game, player uuid.UUID, user uuid.UUID) string {
	t.Helper()
	return s.issue(t, auth.Principal{Role: auth.RolePlayer, GameID: game, PlayerID: player, UserID: user, Name: "seat"})
}

// dial opens a socket with token and query and reads the initial
// snapshot, after which the hub holds the socket (dialAdmitted).
func (s *liveNowStack) dial(t *testing.T, token, query string) {
	t.Helper()
	u := "ws" + strings.TrimPrefix(s.srv.URL, "http") + "/ws?token=" + token
	if query != "" {
		u += "&" + query
	}
	conn, resp, err := websocket.DefaultDialer.Dial(u, nil)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		t.Fatalf("ws dial %s: %v", query, err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, _, err := conn.ReadMessage(); err != nil {
		t.Fatalf("read initial snapshot: %v", err)
	}
}

// live asks GET /admin/live with token and returns the raw body.
func (s *liveNowStack) live(t *testing.T, token string) []byte {
	t.Helper()
	resp := doGet(t, s.srv, "/admin/live", token)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /admin/live: %d %s", resp.StatusCode, raw)
	}
	return raw
}

func decodeLive(t *testing.T, raw []byte) liveNowResponse {
	t.Helper()
	var out liveNowResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decode: %v\n%s", err, raw)
	}
	return out
}

func liveNowTableNamed(t *testing.T, r liveNowResponse, name string) liveNowTable {
	t.Helper()
	for _, tb := range r.Tables {
		if tb.Name == name {
			return tb
		}
	}
	t.Fatalf("Live now has no table %q", name)
	return liveNowTable{}
}

func startWithDecks(t *testing.T, l *Lobby, meta GameMeta, players ...uuid.UUID) {
	t.Helper()
	for _, p := range players {
		if _, err := l.SetDeck(meta.ID, p, "deck", botDeck(20)); err != nil {
			t.Fatalf("SetDeck: %v", err)
		}
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
}

// ADR 0124 §8: a table with two human seats (a connected guest and a
// signed-in person with no socket), a bot and a connected agent, a
// spectator, an admin, a practice table and a waiting table. The
// totals Live now serves are what the tables collector reports over
// the same lobby and hub. With no database, names come from memory.
func TestAdminLiveTotalsMatchTheTiles(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	s := newLiveNowStack(t, nil)
	l := s.lobby

	meta, err := l.Create("Running")
	if err != nil {
		t.Fatal(err)
	}
	_, alice, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	_, bob, err := l.JoinAs(meta.ID, meta.InviteToken, "Bob", DiscordIdentity{}, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	_, agent, err := l.JoinAgent(meta.ID, meta.InviteToken, "Claude", AgentDecl{Client: "Claude Code"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.AddBot(meta.ID, "Bot 1", "random", "", "Mono Red", botDeck(20)); err != nil {
		t.Fatal(err)
	}
	startWithDecks(t, l, meta, alice, bob, agent)
	s.dial(t, s.seatToken(t, meta.ID, alice, uuid.Nil), "")
	s.dial(t, s.seatToken(t, meta.ID, agent, uuid.Nil), "")
	spec, err := spectateOverHTTP(s.srv, meta)
	if err != nil {
		t.Fatal(err)
	}
	s.dial(t, spec, "")
	admin := adminToken(t, s.srv)
	s.dial(t, admin, "game="+meta.ID.String())

	human, pbot := practiceSeats("user:p")
	if _, _, err := l.CreatePractice(human, pbot); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Create("Waiting"); err != nil {
		t.Fatal(err)
	}

	got := decodeLive(t, s.live(t, admin))
	tiles := series(t, s.reg)
	players := 0
	for k, v := range tiles {
		if strings.HasPrefix(k, "cmdctrl_seats_connected{") {
			players += int(v)
		}
	}
	for name, pair := range map[string][2]int{
		"players_connected": {got.Totals.PlayersConnected, players},
		"spectators":        {got.Totals.Spectators, int(tiles["cmdctrl_spectators_connected{}"])},
		"bot_seats":         {got.Totals.BotSeats, int(tiles["cmdctrl_seats{kind=bot}"])},
		"practice_tables":   {got.Totals.PracticeTables, int(tiles["cmdctrl_practice_games{}"])},
	} {
		if pair[0] != pair[1] {
			t.Errorf("totals.%s = %d, the tile reports %d", name, pair[0], pair[1])
		}
	}
	// And the numbers themselves, so two equal wrong answers fail too.
	want := liveNowTotals{PlayersConnected: 2, Spectators: 1, BotSeats: 1, PracticeTables: 1, AdminViews: 1}
	if got.Totals != want {
		t.Errorf("totals = %+v, want %+v", got.Totals, want)
	}
	if got.UnboundSockets != 0 {
		t.Errorf("unbound_sockets = %d, want 0", got.UnboundSockets)
	}

	// The running table and the practice table (running, no socket)
	// are listed; the waiting table with no socket is not.
	if len(got.Tables) != 2 {
		t.Errorf("listed %d tables, want the running and the practice table", len(got.Tables))
	}
	r := liveNowTableNamed(t, got, "Running")
	kinds := map[string]liveNowSeat{}
	for _, seat := range r.Seats {
		kinds[seat.Kind+"/"+cmpName(seat)] = seat
	}
	if a := kinds["human/Alice"]; a.Connected != 1 || a.Since == 0 || a.Account != nil {
		t.Errorf("Alice = %+v, want a connected guest", a)
	}
	if b := kinds["human/Bob"]; b.Connected != 0 || b.Since != 0 || b.Account == nil || b.Account.Name != "Bob" {
		t.Errorf("Bob = %+v, want an account named from the seat, not connected", b)
	}
	if c := kinds["agent/Claude"]; c.Connected != 1 || c.AgentClient != "claude-code" {
		t.Errorf("agent = %+v", c)
	}
	if b := kinds["bot/Bot 1"]; b.BotTier != "random" || b.Connected != 0 || b.DeckName != "Mono Red" {
		t.Errorf("bot = %+v", b)
	}
	if len(r.Spectators) != 1 || r.Spectators[0].Account != nil || r.Spectators[0].Since == 0 {
		t.Errorf("spectators = %+v, want one guest", r.Spectators)
	}
	if len(r.Admins) != 1 || r.Admins[0].Account != nil || r.Admins[0].AsSeat != nil {
		t.Errorf("admins = %+v, want the token, at no seat", r.Admins)
	}
}

func cmpName(s liveNowSeat) string {
	if s.Account != nil {
		return s.Account.Name
	}
	return s.GuestName
}

// liveFields is every key path GET /admin/live may serve (ADR 0124
// §3.4, §8). A new field is a deliberate edit here.
var liveFields = map[string]bool{
	"generated_at": true, "unbound_sockets": true,
	"tables": true, "tables[]": true,
	"tables[].id": true, "tables[].name": true, "tables[].state": true, "tables[].practice": true, "tables[].archived": true,
	"tables[].seats": true, "tables[].seats[]": true,
	"tables[].seats[].seat": true, "tables[].seats[].kind": true,
	"tables[].seats[].account": true, "tables[].seats[].account.id": true,
	"tables[].seats[].account.name": true, "tables[].seats[].account.avatar_url": true,
	"tables[].seats[].guest_name": true, "tables[].seats[].discord_pending": true,
	"tables[].seats[].bot_tier": true, "tables[].seats[].agent_client": true, "tables[].seats[].deck_name": true,
	"tables[].seats[].host": true, "tables[].seats[].connected": true, "tables[].seats[].since": true,
	"tables[].spectators": true, "tables[].spectators[]": true,
	"tables[].spectators[].account": true, "tables[].spectators[].account.id": true,
	"tables[].spectators[].account.name": true, "tables[].spectators[].account.avatar_url": true,
	"tables[].spectators[].since": true,
	"tables[].admins":             true, "tables[].admins[]": true,
	"tables[].admins[].account": true, "tables[].admins[].account.id": true,
	"tables[].admins[].account.name": true, "tables[].admins[].account.avatar_url": true,
	"tables[].admins[].as_seat": true, "tables[].admins[].since": true,
	"totals": true, "totals.players_connected": true, "totals.spectators": true, "totals.bot_seats": true,
	"totals.practice_tables": true, "totals.admin_views": true,
}

// forbiddenKeyParts may not appear in any key (ADR 0124 §8).
var forbiddenKeyParts = []string{"token", "secret", "refresh", "password", "ip", "remote", "addr", "invite",
	"subject", "snowflake", "discord_id", "scope", "email", "requester"}

// walkJSON calls visit with every key path in v and, for each string
// value, its path and the string.
func walkJSON(v any, path string, keys map[string]bool, values func(path, s string)) {
	switch x := v.(type) {
	case map[string]any:
		for k, vv := range x {
			p := k
			if path != "" {
				p = path + "." + k
			}
			keys[p] = true
			walkJSON(vv, p, keys, values)
		}
	case []any:
		keys[path+"[]"] = true
		for _, vv := range x {
			walkJSON(vv, path+"[]", keys, values)
		}
	case string:
		values(path, x)
	}
}

// TestAdminLiveServesOnlyItsFields seeds a database with accounts, a
// sealed refresh token, a pending Discord seat and invites, connects
// seats, a signed-in and a guest spectator and two admin sockets, and
// walks the JSON: every key on the pinned list, no key naming a
// secret, the snowflakes only inside avatar_url and the pending one
// nowhere, no invite or session token anywhere. Then what it says.
func TestAdminLiveServesOnlyItsFields(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	ctx := context.Background()
	d, err := db.Open(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	sealer, err := users.NewSealer(strings.Repeat("k", 40))
	if err != nil {
		t.Fatal(err)
	}
	us := users.NewSQLStore(d, sealer)
	s := newLiveNowStack(t, func(c *Config) { c.Users = us })
	l := s.lobby

	const (
		bobSnowflake   = "100000000000000001"
		carolSnowflake = "100000000000000002"
		daveSnowflake  = "100000000000000003"
		refreshToken   = "refresh-token-never-served"
	)
	bob, err := us.UpsertFromDiscord(ctx, discord.User{ID: bobSnowflake, Username: "bob", GlobalName: "Bobby", Avatar: "bobhash"}, refreshToken, "identify")
	if err != nil {
		t.Fatal(err)
	}
	carol, err := us.UpsertFromDiscord(ctx, discord.User{ID: carolSnowflake, Username: "carol", GlobalName: "Carol", Avatar: "carolhash"}, refreshToken, "identify")
	if err != nil {
		t.Fatal(err)
	}

	meta, err := l.Create("Live")
	if err != nil {
		t.Fatal(err)
	}
	_, alice, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	_, bobSeat, err := l.JoinAs(meta.ID, meta.InviteToken, "Bob",
		DiscordIdentity{ID: bobSnowflake, Username: "bob", GlobalName: "Bobby", AvatarHash: "bobhash"}, bob.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, dave, err := l.JoinAs(meta.ID, meta.InviteToken, "Dave",
		DiscordIdentity{ID: daveSnowflake, Username: "dave", GlobalName: "Dave"}, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.AddBot(meta.ID, "Bot 1", "random", "", "Mono Red", botDeck(20)); err != nil {
		t.Fatal(err)
	}
	startWithDecks(t, l, meta, alice, bobSeat, dave)

	waiting, err := l.Create("Waiting with an agent")
	if err != nil {
		t.Fatal(err)
	}
	_, agent, err := l.JoinAgent(waiting.ID, waiting.InviteToken, "Claude", AgentDecl{Client: "Claude Code"})
	if err != nil {
		t.Fatal(err)
	}

	aliceTok := s.seatToken(t, meta.ID, alice, uuid.Nil)
	bobTok := s.seatToken(t, meta.ID, bobSeat, bob.ID)
	carolTok := s.issue(t, auth.Principal{Role: auth.RoleSpectator, GameID: meta.ID, UserID: carol.ID, DiscordID: carolSnowflake, Name: "Carol"})
	guestSpec, err := spectateOverHTTP(s.srv, meta)
	if err != nil {
		t.Fatal(err)
	}
	admin := adminToken(t, s.srv)
	agentTok := s.seatToken(t, waiting.ID, agent, uuid.Nil)
	s.dial(t, aliceTok, "")
	s.dial(t, bobTok, "")
	s.dial(t, carolTok, "")
	s.dial(t, guestSpec, "")
	s.dial(t, admin, "game="+meta.ID.String())
	s.dial(t, admin, "game="+meta.ID.String()+"&player="+alice.String())
	s.dial(t, agentTok, "")
	if n := s.hub.Count(); n != 7 {
		t.Fatalf("hub has %d sockets, want 7", n)
	}

	raw := s.live(t, admin)
	var body any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	keys := map[string]bool{}
	var leaks []string
	secrets := map[string]string{
		"a player invite":              meta.InviteToken,
		"a spectator invite":           meta.SpectatorInvite,
		"an invite":                    waiting.InviteToken,
		"the refresh token":            refreshToken,
		"the pending seat's snowflake": daveSnowflake,
	}
	for name, tok := range map[string]string{"alice": aliceTok, "bob": bobTok, "carol": carolTok, "guest spectator": guestSpec, "admin": admin, "agent": agentTok} {
		secrets[name+"'s session token"] = tok
	}
	walkJSON(body, "", keys, func(path, v string) {
		for name, secret := range secrets {
			if secret != "" && strings.Contains(v, secret) {
				leaks = append(leaks, fmt.Sprintf("%s at %s", name, path))
			}
		}
		for _, sf := range []string{bobSnowflake, carolSnowflake} {
			if strings.Contains(v, sf) && !strings.HasSuffix(path, ".avatar_url") {
				leaks = append(leaks, fmt.Sprintf("a snowflake outside avatar_url, at %s", path))
			}
		}
	})
	for _, leak := range leaks {
		t.Errorf("GET /admin/live serves %s", leak)
	}
	var paths []string
	for k := range keys {
		paths = append(paths, k)
	}
	sort.Strings(paths)
	for _, p := range paths {
		if !liveFields[p] {
			t.Errorf("GET /admin/live serves %q, which is not on its field list", p)
		}
		parts := strings.Split(strings.ReplaceAll(p, "[]", ""), ".")
		for _, part := range parts {
			for _, bad := range forbiddenKeyParts {
				if strings.Contains(part, bad) {
					t.Errorf("GET /admin/live has a key %q (in %q)", part, p)
				}
			}
		}
	}

	// What it says.
	got := decodeLive(t, raw)
	r := liveNowTableNamed(t, got, "Live")
	if r.State != "active" || r.Archived || r.Practice {
		t.Errorf("Live = %+v", r)
	}
	for _, seat := range r.Seats {
		switch {
		case seat.Kind == metrics.SeatBot:
			if seat.BotTier != "random" || seat.Connected != 0 || seat.Account != nil || seat.GuestName != "Bot 1" {
				t.Errorf("bot seat = %+v", seat)
			}
		case seat.Account != nil:
			want := liveNowAccount{ID: bob.ID.String(), Name: "Bobby", AvatarURL: users.AvatarPath(bobSnowflake, "bobhash")}
			if *seat.Account != want || seat.Connected != 1 || seat.DiscordPending {
				t.Errorf("Bob's seat = %+v (account %+v), want %+v connected", seat, *seat.Account, want)
			}
		case seat.DiscordPending:
			if seat.GuestName == "" || seat.Connected != 0 {
				t.Errorf("Dave's seat = %+v", seat)
			}
		default:
			// Alice: her own socket and the admin bound to her seat.
			if seat.GuestName != "Alice" || seat.Connected != 2 || seat.Since == 0 || !seat.Host {
				t.Errorf("Alice's seat = %+v, want the host, connected twice", seat)
			}
		}
	}
	if len(r.Spectators) != 2 {
		t.Fatalf("spectators = %+v, want Carol and a guest", r.Spectators)
	}
	var named, guests int
	for _, sp := range r.Spectators {
		if sp.Account == nil {
			guests++
			continue
		}
		named++
		want := liveNowAccount{ID: carol.ID.String(), Name: "Carol", AvatarURL: users.AvatarPath(carolSnowflake, "carolhash")}
		if *sp.Account != want {
			t.Errorf("Carol = %+v, want %+v from the database", *sp.Account, want)
		}
	}
	if named != 1 || guests != 1 {
		t.Errorf("spectators: %d named and %d guests, want one each", named, guests)
	}
	var atSeat, watching int
	for _, a := range r.Admins {
		if a.Account != nil {
			t.Errorf("the token's admin socket has an account: %+v", a.Account)
		}
		if a.AsSeat == nil {
			watching++
		} else if *a.AsSeat != 0 {
			t.Errorf("admin as_seat = %d, want Alice's seat 0", *a.AsSeat)
		} else {
			atSeat++
		}
	}
	if atSeat != 1 || watching != 1 {
		t.Errorf("admins = %+v, want one at Alice's seat and one watching", r.Admins)
	}

	w := liveNowTableNamed(t, got, "Waiting with an agent")
	if w.State != "lobby" || len(w.Seats) != 1 || w.Seats[0].AgentClient != "claude-code" || w.Seats[0].Connected != 1 {
		t.Errorf("waiting table = %+v", w)
	}
	// Players connected counts running tables only, as the tile does:
	// Alice and Bob, not the agent waiting in the lobby.
	want := liveNowTotals{PlayersConnected: 2, Spectators: 2, BotSeats: 1, AdminViews: 2}
	if got.Totals != want {
		t.Errorf("totals = %+v, want %+v", got.Totals, want)
	}
}

// Live now without the hub wired answers 503, never an empty page.
func TestAdminLiveWithNoHubIs503(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	s := newLiveNowStack(t, func(c *Config) { c.LiveSockets = nil })
	resp := doGet(t, s.srv, "/admin/live", adminToken(t, s.srv))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("GET /admin/live with no hub: %d, want 503", resp.StatusCode)
	}
}

// lobbyFreeSockets is the hub, asserting at each read that the lobby's
// lock is not held: the handler reads the sockets before the tables.
type lobbyFreeSockets struct {
	hub   *ws.Hub
	lobby *Lobby
	held  atomic.Int64
	calls atomic.Int64
}

func (f *lobbyFreeSockets) LiveSockets() []ws.LiveSocket {
	f.calls.Add(1)
	if !f.lobby.mu.TryLock() {
		f.held.Add(1)
	} else {
		f.lobby.mu.Unlock()
	}
	return f.hub.LiveSockets()
}

// lobbyFreeNames is the database, asserting the same at each read:
// the handler has released the lobby before it asks for names.
type lobbyFreeNames struct {
	users.NoStore
	lobby *Lobby
	held  atomic.Int64
	calls atomic.Int64
}

func (f *lobbyFreeNames) Names(context.Context, []uuid.UUID) (map[uuid.UUID]users.User, error) {
	f.calls.Add(1)
	if !f.lobby.mu.TryLock() {
		f.held.Add(1)
	} else {
		f.lobby.mu.Unlock()
	}
	return map[uuid.UUID]users.User{}, nil
}

// The handler holds one lock at a time and no room lock: it answers
// while a commit holds the table's room, reads the sockets and the
// names with the lobby released, and leaves the lobby free.
func TestAdminLiveHoldsOneLockAtATime(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	var sockets *lobbyFreeSockets
	var names *lobbyFreeNames
	s := newLiveNowStack(t, func(c *Config) {
		sockets = &lobbyFreeSockets{hub: c.LiveSockets.(*ws.Hub), lobby: c.Lobby}
		names = &lobbyFreeNames{lobby: c.Lobby}
		c.LiveSockets = sockets
		c.Users = names
	})
	meta, alice, _ := startTwoSeatGame(t, s.lobby, "Locks")
	// A signed-in spectator, so there is a name to look up.
	s.dial(t, s.issue(t, auth.Principal{Role: auth.RoleSpectator, GameID: meta.ID, UserID: uuid.New(), Name: "Carol"}), "")
	admin := adminToken(t, s.srv)

	room := s.lobby.RoomOf(meta.ID)
	errRollback := errors.New("roll back")
	_, _, err := room.Apply(alice, func() error {
		done := make(chan int, 1)
		go func() {
			req, _ := http.NewRequest(http.MethodGet, s.srv.URL+"/admin/live", nil)
			req.Header.Set("Authorization", "Bearer "+admin)
			resp, err := s.srv.Client().Do(req)
			if err != nil {
				done <- 0
				return
			}
			_ = resp.Body.Close()
			done <- resp.StatusCode
		}()
		select {
		case code := <-done:
			if code != http.StatusOK {
				t.Errorf("GET /admin/live inside a commit: %d", code)
			}
		case <-time.After(5 * time.Second):
			t.Error("GET /admin/live blocked on the room's lock")
		}
		return errRollback
	})
	if !errors.Is(err, errRollback) {
		t.Fatalf("Apply: %v", err)
	}
	if sockets.calls.Load() == 0 || names.calls.Load() == 0 {
		t.Fatalf("the handler read the sockets %d times and the names %d times; want both", sockets.calls.Load(), names.calls.Load())
	}
	if n := sockets.held.Load(); n != 0 {
		t.Errorf("the sockets were read %d times with the lobby's lock held", n)
	}
	if n := names.held.Load(); n != 0 {
		t.Errorf("the names were read %d times with the lobby's lock held", n)
	}
	if !s.lobby.mu.TryLock() {
		t.Fatal("GET /admin/live left the lobby's lock held")
	}
	s.lobby.mu.Unlock()
}
