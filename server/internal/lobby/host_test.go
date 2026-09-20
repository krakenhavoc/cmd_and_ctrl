package lobby

// host_test.go: ADR 0075 §2.1, the table host.

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
)

func mustGet(t *testing.T, l *Lobby, id uuid.UUID) GameMeta {
	t.Helper()
	m, err := l.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	return m
}

// assertHost checks the meta, every SeatInfo.IsHost flag and the
// room's game view all name the same host.
func assertHost(t *testing.T, l *Lobby, id, want uuid.UUID) {
	t.Helper()
	m := mustGet(t, l, id)
	if m.HostPlayerID != want {
		t.Errorf("HostPlayerID = %v, want %v", m.HostPlayerID, want)
	}
	for _, s := range m.Players {
		if s.IsHost != (s.PlayerID == want && want != uuid.Nil) {
			t.Errorf("seat %s IsHost = %v (host %v)", s.Name, s.IsHost, want)
		}
	}
	view, _, err := l.RoomOf(id).Snapshot()
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	for _, s := range view.Seats {
		if s.IsHost != (want != uuid.Nil && s.ID == want.String()) {
			t.Errorf("view seat %s is_host = %v (host %v)", s.Name, s.IsHost, want)
		}
	}
}

func TestFirstHumanSeatHosts(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("FNM")
	// A bot sits down first: it never hosts, so the table has none.
	_, botID, err := l.AddBot(meta.ID, "Bot", "random", "", "d", botDeck(5))
	if err != nil {
		t.Fatal(err)
	}
	assertHost(t, l, meta.ID, uuid.Nil)

	_, alice, _ := l.Join(meta.ID, meta.InviteToken, "Alice")
	_, bob, _ := l.Join(meta.ID, meta.InviteToken, "Bob")
	assertHost(t, l, meta.ID, alice)
	if alice == bob || botID == alice {
		t.Fatal("ids collide")
	}
}

func TestBotOnlyTableHasNoHost(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("bots")
	for i := 0; i < 2; i++ {
		if _, _, err := l.AddBot(meta.ID, "Bot", "random", "", "d", botDeck(5)); err != nil {
			t.Fatal(err)
		}
	}
	assertHost(t, l, meta.ID, uuid.Nil)
	for _, s := range mustGet(t, l, meta.ID).Players {
		if s.IsHost {
			t.Errorf("bot seat %s is host", s.Name)
		}
	}
}

func TestNamedHostBindsWhenTheyClaimASeat(t *testing.T) {
	l := newTestLobby(t)
	meta, err := l.CreateWith("FNM", uuid.Nil, "1111")
	if err != nil {
		t.Fatal(err)
	}
	if meta.HostDiscordID != "" {
		t.Errorf("create response leaks the pending host: %q", meta.HostDiscordID)
	}
	// Somebody else arrives first and hosts in the meantime.
	_, alice, _ := l.JoinWithIdentity(meta.ID, meta.InviteToken, "", DiscordIdentity{ID: "2222", Username: "alice"})
	assertHost(t, l, meta.ID, alice)

	// The named host sits down and takes the table.
	_, luke, err := l.JoinWithIdentity(meta.ID, meta.InviteToken, "", DiscordIdentity{ID: "1111", Username: "luke"})
	if err != nil {
		t.Fatal(err)
	}
	assertHost(t, l, meta.ID, luke)

	// The pending name is spent: the host can hand the table on and
	// it stays handed on.
	l.mu.Lock()
	pending := l.games[meta.ID].meta.HostDiscordID
	l.mu.Unlock()
	if pending != "" {
		t.Errorf("pending host still %q after the claim", pending)
	}
	if _, err := l.TransferHost(meta.ID, alice); err != nil {
		t.Fatal(err)
	}
	assertHost(t, l, meta.ID, alice)
}

func TestNamedHostArrivingFirstHostsImmediately(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.CreateWith("FNM", uuid.Nil, "1111")
	_, luke, _ := l.JoinWithIdentity(meta.ID, meta.InviteToken, "", DiscordIdentity{ID: "1111", Username: "luke"})
	_, _, _ = l.JoinWithIdentity(meta.ID, meta.InviteToken, "", DiscordIdentity{ID: "2222", Username: "alice"})
	assertHost(t, l, meta.ID, luke)
}

func TestTransferHostValidatesTheTarget(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("FNM")
	_, alice, _ := l.Join(meta.ID, meta.InviteToken, "Alice")
	_, bob, _ := l.Join(meta.ID, meta.InviteToken, "Bob")
	_, botID, _ := l.AddBot(meta.ID, "Bot", "random", "", "d", botDeck(5))

	if _, err := l.TransferHost(meta.ID, botID); !errors.Is(err, ErrHostIneligible) {
		t.Errorf("transfer to bot: %v, want ErrHostIneligible", err)
	}
	if _, err := l.TransferHost(meta.ID, uuid.New()); !errors.Is(err, ErrPlayerNotInGame) {
		t.Errorf("transfer to stranger: %v, want ErrPlayerNotInGame", err)
	}
	if _, err := l.TransferHost(uuid.New(), bob); !errors.Is(err, ErrGameNotFound) {
		t.Errorf("transfer on unknown game: %v", err)
	}
	assertHost(t, l, meta.ID, alice)

	after, err := l.TransferHost(meta.ID, bob)
	if err != nil {
		t.Fatal(err)
	}
	if after.HostPlayerID != bob {
		t.Errorf("returned meta host %v, want bob", after.HostPlayerID)
	}
	assertHost(t, l, meta.ID, bob)
}

// startedTable seats Alice, a bot, Bob and Carol (in that seat order)
// and starts the game.
func startedTable(t *testing.T, l *Lobby) (GameMeta, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	meta, _ := l.Create("FNM")
	_, alice, _ := l.Join(meta.ID, meta.InviteToken, "Alice")
	_, botID, err := l.AddBot(meta.ID, "Bot", "random", "", "d", botDeck(20))
	if err != nil {
		t.Fatal(err)
	}
	_, bob, _ := l.Join(meta.ID, meta.InviteToken, "Bob")
	_, carol, _ := l.Join(meta.ID, meta.InviteToken, "Carol")
	for _, p := range []uuid.UUID{alice, bob, carol} {
		uploadDummyDeck(t, l, meta.ID, p)
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return meta, alice, botID, bob, carol
}

func concede(t *testing.T, l *Lobby, id, player uuid.UUID) {
	t.Helper()
	room := l.RoomOf(id)
	if _, _, err := room.Apply(player, func() error { return room.Game.Concede(player) }); err != nil {
		t.Fatalf("concede: %v", err)
	}
}

func TestHostPassesInTurnOrderWhenTheHostLeaves(t *testing.T) {
	l := newTestLobby(t)
	meta, alice, _, bob, carol := startedTable(t, l)
	assertHost(t, l, meta.ID, alice)

	// Alice leaves: the next seat is a bot, which never hosts, so
	// hosting skips to Bob.
	concede(t, l, meta.ID, alice)
	assertHost(t, l, meta.ID, bob)

	concede(t, l, meta.ID, bob)
	assertHost(t, l, meta.ID, carol)

	// The last human leaves: nobody hosts; the admin still manages.
	concede(t, l, meta.ID, carol)
	assertHost(t, l, meta.ID, uuid.Nil)
}

func TestHostPassWrapsAroundTheTable(t *testing.T) {
	l := newTestLobby(t)
	meta, alice, _, _, carol := startedTable(t, l)
	if _, err := l.TransferHost(meta.ID, carol); err != nil {
		t.Fatal(err)
	}
	// Carol is the last seat; the next human in turn order is Alice.
	concede(t, l, meta.ID, carol)
	assertHost(t, l, meta.ID, alice)

	// A departed player cannot be handed the table back.
	if _, err := l.TransferHost(meta.ID, carol); !errors.Is(err, ErrHostIneligible) {
		t.Errorf("transfer to eliminated seat: %v, want ErrHostIneligible", err)
	}
}

func TestCanManageTable(t *testing.T) {
	game1, game2 := uuid.New(), uuid.New()
	host, other := uuid.New(), uuid.New()
	meta := GameMeta{ID: game1, HostPlayerID: host}

	cases := []struct {
		name string
		p    auth.Principal
		want bool
	}{
		{"admin", auth.Principal{Role: auth.RoleAdmin}, true},
		{"host", auth.Principal{Role: auth.RolePlayer, GameID: game1, PlayerID: host}, true},
		{"other player", auth.Principal{Role: auth.RolePlayer, GameID: game1, PlayerID: other}, false},
		{"spectator", auth.Principal{Role: auth.RoleSpectator, GameID: game1}, false},
		{"identified", auth.Principal{Role: auth.RoleIdentified, DiscordID: "1"}, false},
		{"host of another game", auth.Principal{Role: auth.RolePlayer, GameID: game2, PlayerID: host}, false},
		{"empty principal", auth.Principal{}, false},
	}
	for _, tc := range cases {
		if got := CanManageTable(tc.p, meta); got != tc.want {
			t.Errorf("%s: CanManageTable = %v, want %v", tc.name, got, tc.want)
		}
	}
	// No host at all: only the admin.
	hostless := GameMeta{ID: game1}
	if CanManageTable(auth.Principal{Role: auth.RolePlayer, GameID: game1}, hostless) {
		t.Error("a player with a nil PlayerID manages a hostless table")
	}
	if !CanManageTable(auth.Principal{Role: auth.RoleAdmin}, hostless) {
		t.Error("admin cannot manage a hostless table")
	}
}

func TestHostSurvivesARestart(t *testing.T) {
	dir := t.TempDir()
	log := quietLogger()
	l, _ := newDurableLobby(t, dir)

	meta, _ := l.Create("FNM")
	_, _, _ = l.Join(meta.ID, meta.InviteToken, "Alice")
	_, bob, _ := l.Join(meta.ID, meta.InviteToken, "Bob")
	if _, err := l.TransferHost(meta.ID, bob); err != nil {
		t.Fatal(err)
	}

	// A second table whose named host has not arrived yet.
	pending, _ := l.CreateWith("Later", uuid.Nil, "1111")
	_, carol, _ := l.Join(pending.ID, pending.InviteToken, "Carol")

	// --- the deploy -----------------------------------------------
	l2, _ := newDurableLobby(t, dir)
	if n := l2.RestoreFromDisk(log); n != 2 {
		t.Fatalf("restored %d games, want 2", n)
	}
	assertHost(t, l2, meta.ID, bob)
	assertHost(t, l2, pending.ID, carol)

	// What the games row says, read straight from the store.
	loadRow := func(id uuid.UUID) GameRecord {
		t.Helper()
		ctx, cancel := storeCtx()
		defer cancel()
		rec, _, err := l2.store.LoadGame(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		return rec
	}
	if rec := loadRow(pending.ID); rec.HostPlayerID != carol || rec.HostDiscordID != "1111" {
		t.Errorf("row before claim: host %v / pending %q, want carol / 1111", rec.HostPlayerID, rec.HostDiscordID)
	}

	// The pending named host survived too: when they arrive, they
	// take the table, and the row records it.
	_, luke, err := l2.JoinWithIdentity(pending.ID, pending.InviteToken, "", DiscordIdentity{ID: "1111", Username: "luke"})
	if err != nil {
		t.Fatal(err)
	}
	assertHost(t, l2, pending.ID, luke)
	if rec := loadRow(pending.ID); rec.HostPlayerID != luke || rec.HostDiscordID != "" {
		t.Errorf("row after claim: host %v / pending %q, want luke / none", rec.HostPlayerID, rec.HostDiscordID)
	}
}

// TestHostPassSurvivesARestart: a pass observed after elimination is
// written to the row, and a restart does not hand the table back.
func TestHostPassSurvivesARestart(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, alice, _, bob, _ := startedTable(t, l)
	concede(t, l, meta.ID, alice)
	assertHost(t, l, meta.ID, bob)

	l2, _ := newDurableLobby(t, dir)
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games, want 1", n)
	}
	assertHost(t, l2, meta.ID, bob)
}

func TestTransferHostRoute(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	meta, _ := l.Create("FNM")
	alice := joinAs(t, srv, meta, "Alice")
	bob := joinAs(t, srv, meta, "Bob")
	_, botID, err := l.AddBot(meta.ID, "Bot", "random", "", "d", botDeck(5))
	if err != nil {
		t.Fatal(err)
	}
	other, _ := l.Create("Other")
	otherHost := joinAs(t, srv, other, "Dana")
	admin := adminSession(t, a)
	path := "/games/" + meta.ID.String() + "/host"

	post := func(token string, target uuid.UUID) int {
		t.Helper()
		resp := postJSON(t, srv, path, token, transferHostRequest{PlayerID: target})
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)
		return resp.StatusCode
	}

	if got := post(bob.Token, bob.PlayerID); got != http.StatusForbidden {
		t.Errorf("non-host player: %d, want 403", got)
	}
	if got := post(otherHost.Token, bob.PlayerID); got != http.StatusForbidden {
		t.Errorf("host of another table: %d, want 403", got)
	}
	if got := post(alice.Token, botID); got != http.StatusUnprocessableEntity {
		t.Errorf("to a bot: %d, want 422", got)
	}
	if got := post(alice.Token, uuid.New()); got != http.StatusUnprocessableEntity {
		t.Errorf("to an unseated player: %d, want 422", got)
	}
	if got := post("", bob.PlayerID); got != http.StatusUnauthorized {
		t.Errorf("anonymous: %d, want 401", got)
	}
	assertHost(t, l, meta.ID, alice.PlayerID)

	if got := post(alice.Token, bob.PlayerID); got != http.StatusOK {
		t.Fatalf("host transfers: %d, want 200", got)
	}
	assertHost(t, l, meta.ID, bob.PlayerID)

	// Alice is no longer host, so she can no longer transfer.
	if got := post(alice.Token, alice.PlayerID); got != http.StatusForbidden {
		t.Errorf("former host: %d, want 403", got)
	}
	if got := post(admin, alice.PlayerID); got != http.StatusOK {
		t.Errorf("admin transfers: %d, want 200", got)
	}
	assertHost(t, l, meta.ID, alice.PlayerID)
}

func TestCreateRouteAcceptsANamedHost(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	admin := adminSession(t, a)
	resp := postJSON(t, srv, "/games", admin, createGameRequest{Name: "FNM", HostDiscordID: "1111"})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: %d", resp.StatusCode)
	}
	list := l.List()
	if len(list) != 1 {
		t.Fatalf("games: %d", len(list))
	}
	id := list[0].ID
	l.mu.Lock()
	pending := l.games[id].meta.HostDiscordID
	l.mu.Unlock()
	if pending != "1111" {
		t.Errorf("pending host = %q, want 1111", pending)
	}
}
