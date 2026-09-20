package lobby

// spawn_test.go — ADR 0075 §2.4, the production spawner.
//
// The gates are the whole feature, so they are tested from the
// outside, through the HTTP stack, with real sessions: a unit test on
// CanManageTable would pass with the route wired to nothing.

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/appenv"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

type spawnFixture struct {
	t      *testing.T
	srv    *httptest.Server
	lobby  *Lobby
	gameID uuid.UUID
	// alice hosts (first human seat); bob is an ordinary seat.
	alice, bob             uuid.UUID
	aliceTok, bobTok       string
	adminTok, spectatorTok string
}

// newSpawnFixture starts a two-seat game on a PRODUCTION stack — the
// spawn route is not dev-gated, so the fixture must not be either —
// with the card index and the token templates wired.
func newSpawnFixture(t *testing.T) *spawnFixture {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	l := NewLobby(mgr)
	a := auth.NewMemoryAuthenticator()
	srv := httptest.NewServer(Handler(Config{
		Lobby:      l,
		Auth:       a,
		AdminToken: "shared-admin-token",
		Cards:      devIndex(t),
		Tokens:     effects.Tokens(),
		Env:        appenv.EnvProd,
		Features:   appenv.LoadFeatures(appenv.EnvProd),
	}))
	t.Cleanup(srv.Close)

	meta, err := l.Create("Spawn")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, alice, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatalf("join Alice: %v", err)
	}
	_, bob, err := l.Join(meta.ID, meta.InviteToken, "Bob")
	if err != nil {
		t.Fatalf("join Bob: %v", err)
	}
	uploadDummyDeck(t, l, meta.ID, alice)
	uploadDummyDeck(t, l, meta.ID, bob)
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if got := mustGet(t, l, meta.ID).HostPlayerID; got != alice {
		t.Fatalf("host = %v, want Alice %v", got, alice)
	}

	issue := func(p auth.Principal) string {
		tok, _, err := a.Issue(context.Background(), p, time.Hour)
		if err != nil {
			t.Fatalf("issue: %v", err)
		}
		return tok
	}
	return &spawnFixture{
		t: t, srv: srv, lobby: l, gameID: meta.ID, alice: alice, bob: bob,
		aliceTok: issue(auth.Principal{Role: auth.RolePlayer, GameID: meta.ID, PlayerID: alice, Name: "Alice"}),
		bobTok:   issue(auth.Principal{Role: auth.RolePlayer, GameID: meta.ID, PlayerID: bob, Name: "Bob"}),
		adminTok: issue(auth.Principal{Role: auth.RoleAdmin}),
		spectatorTok: issue(auth.Principal{
			Role: auth.RoleSpectator, GameID: meta.ID, Name: "Watcher",
		}),
	}
}

// allowSpawn flips the table's AllowSpawn setting through the lobby,
// the way the host's PATCH /games/{id}/settings does (sub-PR 3). Going
// through Lobby.UpdateSettings rather than poking Game.UpdateSettings
// keeps the fixture on the surface production uses — the commit lands
// in the replay and the `settings` log line is emitted — so a spawn
// test is set up by the same call a host makes.
func (f *spawnFixture) allowSpawn(on bool) {
	f.t.Helper()
	if _, err := f.lobby.UpdateSettings(f.gameID, f.alice, game.SettingsPatch{AllowSpawn: &on}); err != nil {
		f.t.Fatalf("UpdateSettings: %v", err)
	}
}

func (f *spawnFixture) post(token, body string) *http.Response {
	f.t.Helper()
	return req(f.t, f.srv, http.MethodPost, "/games/"+f.gameID.String()+"/spawn", token, body)
}

// battlefield counts the permanents one seat controls.
func (f *spawnFixture) battlefield(owner uuid.UUID) int {
	f.t.Helper()
	n := 0
	g := f.lobby.RoomOf(f.gameID).Game
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner {
				n++
			}
		}
	})
	return n
}

func (f *spawnFixture) body(extra string) string {
	return `{"player_id":"` + f.alice.String() + `","zone":"battlefield"` + extra + `}`
}

func errorMessage(t *testing.T, resp *http.Response) string {
	t.Helper()
	var payload struct {
		Error string `json:"error"`
		Msg   string `json:"message"`
	}
	raw, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(raw, &payload)
	return payload.Error + payload.Msg + string(raw)
}

// The two gates, and the fact that a refusal says WHICH one fired:
// "you are not the host" and "this table has spawning off" are
// different problems with different fixes.
func TestProductionSpawnNeedsBothGates(t *testing.T) {
	f := newSpawnFixture(t)
	before := f.battlefield(f.alice)

	// Setting off, host calling.
	resp := f.post(f.aliceTok, f.body(`,"name":"Lightning Bolt"`))
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("host with the setting off: status %d, want 403", resp.StatusCode)
	}
	if msg := errorMessage(t, resp); !strings.Contains(msg, "switched off") {
		t.Errorf("message %q does not say the setting is what refused", msg)
	}
	if got := f.battlefield(f.alice); got != before {
		t.Errorf("a refused spawn put %d cards on the battlefield", got-before)
	}

	f.allowSpawn(true)

	// Setting on, a seated NON-host calling.
	resp = f.post(f.bobTok, f.body(`,"name":"Lightning Bolt"`))
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("non-host with the setting on: status %d, want 403", resp.StatusCode)
	}
	if msg := errorMessage(t, resp); !strings.Contains(msg, "host") {
		t.Errorf("message %q does not say the caller is what refused", msg)
	}
	if got := f.battlefield(f.alice); got != before {
		t.Errorf("a refused spawn put %d cards on the battlefield", got-before)
	}

	// A spectator is not a seat at all.
	if got := f.post(f.spectatorTok, f.body(`,"name":"Lightning Bolt"`)).StatusCode; got != http.StatusForbidden {
		t.Errorf("spectator: status %d, want 403", got)
	}

	// Both gates open.
	resp = f.post(f.aliceTok, f.body(`,"name":"Lightning Bolt","count":2`))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("host with the setting on: status %d (%s), want 200", resp.StatusCode, errorMessage(t, resp))
	}
	if got := f.battlefield(f.alice); got != before+2 {
		t.Errorf("battlefield grew by %d, want 2", got-before)
	}
}

// The admin manages every table, with or without a seat.
func TestAdminCanSpawn(t *testing.T) {
	f := newSpawnFixture(t)
	f.allowSpawn(true)
	before := f.battlefield(f.bob)
	body := `{"player_id":"` + f.bob.String() + `","zone":"battlefield","name":"Lightning Bolt"}`
	if resp := f.post(f.adminTok, body); resp.StatusCode != http.StatusOK {
		t.Fatalf("admin spawn: status %d (%s)", resp.StatusCode, errorMessage(t, resp))
	}
	if got := f.battlefield(f.bob); got != before+1 {
		t.Errorf("bob's battlefield grew by %d, want 1", got-before)
	}
}

// Tokens are why the owner wanted this in production: the dev
// spawner resolves a Scryfall printing and a token has none.
func TestProductionSpawnMakesTokens(t *testing.T) {
	f := newSpawnFixture(t)
	f.allowSpawn(true)

	resp := f.post(f.aliceTok, f.body(`,"token":"Treasure","count":3`))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d (%s)", resp.StatusCode, errorMessage(t, resp))
	}
	var out spawnResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !out.Token || out.Name != "Treasure" || out.Count != 3 {
		t.Fatalf("response = %+v, want 3 Treasure tokens", out)
	}
	if out.ScryfallID != "" {
		t.Errorf("scryfall_id = %q, want empty for a token", out.ScryfallID)
	}

	g := f.lobby.RoomOf(f.gameID).Game
	found := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Treasure" && c.IsToken() {
				found++
			}
		}
	})
	if found != 3 {
		t.Errorf("found %d Treasures on the battlefield, want 3", found)
	}

	// CR 704.5d: anywhere else, a token would cease to exist at the
	// next state-based check, so the route refuses rather than
	// pretending.
	bad := `{"player_id":"` + f.alice.String() + `","zone":"hand","token":"Treasure"}`
	resp = f.post(f.aliceTok, bad)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("token into hand: status %d, want 400", resp.StatusCode)
	}
	if msg := errorMessage(t, resp); !strings.Contains(msg, "battlefield") {
		t.Errorf("message %q does not explain the restriction", msg)
	}

	if resp := f.post(f.aliceTok, f.body(`,"token":"No Such Token"`)); resp.StatusCode != http.StatusNotFound {
		t.Errorf("unknown token key: status %d, want 404", resp.StatusCode)
	}
}

func TestSpawnTokenListIsGatedLikeTheSpawn(t *testing.T) {
	f := newSpawnFixture(t)
	path := "/games/" + f.gameID.String() + "/spawn/tokens"

	if got := req(t, f.srv, http.MethodGet, path, f.bobTok, "").StatusCode; got != http.StatusForbidden {
		t.Errorf("non-host: status %d, want 403", got)
	}
	resp := req(t, f.srv, http.MethodGet, path, f.aliceTok, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("host: status %d", resp.StatusCode)
	}
	var out spawnTokenList
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Both halves of the library: a behaviour token and a table row.
	wantAll := map[string]bool{"Treasure": false, "1/1 white Soldier": false}
	for _, k := range out.Tokens {
		if _, ok := wantAll[k]; ok {
			wantAll[k] = true
		}
	}
	for k, seen := range wantAll {
		if !seen {
			t.Errorf("token list does not offer %q", k)
		}
	}
	// The list is what the `token` field accepts, so every key on it
	// must spawn.
	if len(out.Tokens) < 50 {
		t.Errorf("token list has %d entries, want the whole table", len(out.Tokens))
	}
}

// A mistaken spawn is taken back with the ordinary undo, and it is
// FREE: fixing a typo must not cost the host the one take-back they
// get per turn.
func TestProductionSpawnIsUndoableAndFree(t *testing.T) {
	f := newSpawnFixture(t)
	f.allowSpawn(true)
	room := f.lobby.RoomOf(f.gameID)

	before := f.battlefield(f.alice)
	budget := undosRemaining(t, room.Game, f.alice)
	if budget <= 0 {
		t.Fatalf("Alice starts with %d undos; the test needs a budget to watch", budget)
	}

	if resp := f.post(f.aliceTok, f.body(`,"token":"Treasure","count":2`)); resp.StatusCode != http.StatusOK {
		t.Fatalf("spawn: status %d (%s)", resp.StatusCode, errorMessage(t, resp))
	}
	if got := f.battlefield(f.alice); got != before+2 {
		t.Fatalf("battlefield grew by %d, want 2", got-before)
	}

	if _, _, err := room.Undo(f.alice); err != nil {
		t.Fatalf("Undo by the spawner: %v", err)
	}
	if got := f.battlefield(f.alice); got != before {
		t.Errorf("after undo the battlefield is %d, want %d", got, before)
	}
	if got := undosRemaining(t, room.Game, f.alice); got != budget {
		t.Errorf("undos remaining = %d, want %d — a FreeUndo entry must not debit the budget", got, budget)
	}
}

func undosRemaining(t *testing.T, g *game.Game, player uuid.UUID) int {
	t.Helper()
	n := 0
	g.ReadSnapshot(func() {
		for _, p := range g.Seats {
			if p.ID == player {
				n = p.UndosRemaining
			}
		}
	})
	return n
}

// The line the whole feature is conditional on. It names the spawner
// as host — which only the ROOM can know — and it redacts the card
// for a spawn into a hidden zone.
func TestProductionSpawnIsAnnounced(t *testing.T) {
	f := newSpawnFixture(t)
	f.allowSpawn(true)
	body := `{"player_id":"` + f.bob.String() + `","zone":"battlefield","token":"Treasure","count":2}`
	if resp := f.post(f.aliceTok, body); resp.StatusCode != http.StatusOK {
		t.Fatalf("spawn: status %d (%s)", resp.StatusCode, errorMessage(t, resp))
	}

	line := lastSpawnLine(t, f)
	want := "Alice (host) spawned 2 × Treasure onto Bob's battlefield"
	if line != want {
		t.Errorf("log line = %q, want %q", line, want)
	}

	hidden := `{"player_id":"` + f.bob.String() + `","zone":"hand","name":"Lightning Bolt"}`
	if resp := f.post(f.aliceTok, hidden); resp.StatusCode != http.StatusOK {
		t.Fatalf("hidden-zone spawn: status %d (%s)", resp.StatusCode, errorMessage(t, resp))
	}
	line = lastSpawnLine(t, f)
	if strings.Contains(line, "Lightning Bolt") {
		t.Errorf("log line names a card spawned into a hand: %q", line)
	}
	if !strings.Contains(line, "Bob's hand") {
		t.Errorf("log line = %q, want it to name the zone", line)
	}
}

func lastSpawnLine(t *testing.T, f *spawnFixture) string {
	t.Helper()
	view, _, err := f.lobby.RoomOf(f.gameID).Snapshot()
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	line := ""
	for _, e := range view.Log {
		if e.Kind == protocol.LogSpawn {
			line = e.Text
		}
	}
	if line == "" {
		t.Fatal("no spawn entry in the public log")
	}
	return line
}

// The dev route is unchanged by ADR 0075: still dev-only, still open
// to anyone at the table, still needs no setting — and still records
// no undo entry, because a preview box's spawns are not repairs.
func TestDevSpawnRouteIsUnchanged(t *testing.T) {
	srv, l, a := newDevStack(t, appenv.EnvDev, appenv.LoadFeatures(appenv.EnvDev), devIndex(t))
	meta, err := l.Create("dev")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, alice, _ := l.Join(meta.ID, meta.InviteToken, "Alice")
	_, bob, _ := l.Join(meta.ID, meta.InviteToken, "Bob")
	uploadDummyDeck(t, l, meta.ID, alice)
	uploadDummyDeck(t, l, meta.ID, bob)
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// Bob is not the host and the table has never allowed spawning.
	tok, _, err := a.Issue(context.Background(), auth.Principal{
		Role: auth.RolePlayer, GameID: meta.ID, PlayerID: bob, Name: "Bob",
	}, time.Hour)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	body := `{"player_id":"` + bob.String() + `","zone":"battlefield","name":"Lightning Bolt"}`
	resp := req(t, srv, http.MethodPost, "/games/"+meta.ID.String()+"/dev/spawn", tok, body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("dev spawn by a non-host with the setting off: status %d (%s)", resp.StatusCode, errorMessage(t, resp))
	}

	// No undo entry: the dev route still commits through
	// ApplyExternal.
	if _, _, err := l.RoomOf(meta.ID).Undo(bob); !ws.IsErrNothingToUndo(err) {
		t.Errorf("Undo after a dev spawn = %v, want ErrNothingToUndo", err)
	}

	// And it still has no token vocabulary — `token` is a field of
	// the production request only.
	tokenBody := `{"player_id":"` + bob.String() + `","zone":"battlefield","token":"Treasure"}`
	if got := req(t, srv, http.MethodPost, "/games/"+meta.ID.String()+"/dev/spawn", tok, tokenBody).StatusCode; got != http.StatusBadRequest {
		t.Errorf("dev spawn with a token key: status %d, want 400", got)
	}
}

// The production route is NOT dev-gated — that is the amendment to
// ADR 0023 — so it must answer on a production stack rather than 404.
func TestProductionSpawnRouteExistsInProduction(t *testing.T) {
	f := newSpawnFixture(t)
	if got := f.post("", f.body(`,"name":"Lightning Bolt"`)).StatusCode; got != http.StatusUnauthorized {
		t.Fatalf("unauthenticated: status %d, want 401 (the route must exist in production)", got)
	}
}

// GET /dev/cards 404s in production, so the production spawner needs
// a search of its own or the name field has nothing behind it.
func TestSpawnCardSearchIsGatedLikeTheSpawn(t *testing.T) {
	f := newSpawnFixture(t)
	path := "/games/" + f.gameID.String() + "/spawn/cards?q=lightning"

	if got := req(t, f.srv, http.MethodGet, path, f.bobTok, "").StatusCode; got != http.StatusForbidden {
		t.Errorf("non-host: status %d, want 403", got)
	}
	resp := req(t, f.srv, http.MethodGet, path, f.aliceTok, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("host: status %d (%s)", resp.StatusCode, errorMessage(t, resp))
	}
	var out struct {
		Cards []devCardResult `json:"cards"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Cards) != 1 || out.Cards[0].Name != "Lightning Bolt" {
		t.Fatalf("results = %+v, want one Lightning Bolt", out.Cards)
	}
	// And the dev-only search is still 404 on this production stack,
	// which is what makes the sibling necessary rather than redundant.
	if got := req(t, f.srv, http.MethodGet, "/dev/cards?q=lightning", f.adminTok, "").StatusCode; got != http.StatusNotFound {
		t.Errorf("/dev/cards in production: status %d, want 404", got)
	}
}
