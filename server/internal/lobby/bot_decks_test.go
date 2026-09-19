package lobby

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// bot_decks_test.go covers the player-facing half of the Add-bot
// flow: the picker's options endpoint, and seating a bot by NAMED
// deck rather than by pasting a decklist.

// fakeDeckSource is a two-deck catalog whose lists resolve against
// buildBotDeckIndex below. Standing in for both the placeholder and
// (from sub-PR 5) the curated registry — the lobby only ever sees
// the aiseat.DeckSource interface.
type fakeDeckSource struct{}

func (fakeDeckSource) List() []aiseat.DeckInfo {
	return []aiseat.DeckInfo{
		{ID: "test-mono-white", Name: "Mono-white test", Colors: []string{"W"}, Commander: "Test Commander"},
	}
}

func (f fakeDeckSource) Decklist(id string) (aiseat.DeckInfo, string, bool) {
	if id != "test-mono-white" {
		return aiseat.DeckInfo{}, "", false
	}
	return f.List()[0], "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n", true
}

func newBotHTTPStack(t *testing.T, idx *cards.Index, decks aiseat.DeckSource) (*httptest.Server, *Lobby, *fakeBotHost, auth.Authenticator) {
	t.Helper()
	log := discardLogger()
	mgr := ws.NewRoomManager(log, "")
	l := NewLobby(mgr)
	host := newFakeBotHost()
	l.SetBotHost(host)
	a := newTestAuth()
	srv := newTestServerWithConfig(t, Config{
		Lobby: l, Auth: a, AdminToken: "shared-admin-token",
		Cards: idx, Bots: host, BotDecks: decks,
	})
	return srv, l, host, a
}

// joinAs claims a seat and returns the minted session.
func joinAs(t *testing.T, srv *httptest.Server, meta GameMeta, name string) sessionResponse {
	t.Helper()
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: name})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("join %s: got %d (body=%s)", name, resp.StatusCode, body)
	}
	var out sessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestBotOptionsListsEveryTierAndTheDeckCatalog(t *testing.T) {
	srv, _, _, a := newBotHTTPStack(t, buildMinimalDeckIndex(t), fakeDeckSource{})
	token := adminSession(t, a)

	resp := doGet(t, srv, "/bot/options", token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("options: got %d (body=%s)", resp.StatusCode, body)
	}
	var out botOptionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if !out.Enabled {
		t.Error("enabled should be true when a bot host is configured")
	}
	// Every declared tier is listed, not only the buildable one — the
	// picker greys the rest out rather than pretending the difficulty
	// slider has a single notch.
	if len(out.Tiers) != len(aiseat.Tiers()) {
		t.Fatalf("tiers: %+v", out.Tiers)
	}
	available := 0
	for _, ti := range out.Tiers {
		if ti.Available {
			available++
			if ti.Tier != aiseat.TierRandom {
				t.Errorf("unexpected available tier %q", ti.Tier)
			}
		}
	}
	if available != 1 {
		t.Errorf("available tiers: %d, want 1", available)
	}
	if len(out.Decks) != 1 || out.Decks[0].ID != "test-mono-white" {
		t.Errorf("decks: %+v", out.Decks)
	}
}

func TestBotOptionsReportsDisabledWithoutAHost(t *testing.T) {
	srv, _, a, _ := newTestHTTPStackWithCards(t, buildMinimalDeckIndex(t))
	token := adminSession(t, a)
	resp := doGet(t, srv, "/bot/options", token)
	defer resp.Body.Close()
	var out botOptionsResponse
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if out.Enabled {
		t.Error("enabled should be false with no bot host")
	}
	for _, ti := range out.Tiers {
		if ti.Available {
			t.Errorf("tier %q reported available on a server with no bot host", ti.Tier)
		}
	}
}

// The player-facing path: pick a tier and a curated deck, get a seat.
func TestAddBotByNamedDeck(t *testing.T) {
	srv, l, _, _ := newBotHTTPStack(t, buildMinimalDeckIndex(t), fakeDeckSource{})
	meta, _ := l.Create("FNM")
	joined := joinAs(t, srv, meta, "Alice")

	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/seats/bot", joined.Token,
		addBotRequest{Tier: "random", Deck: "test-mono-white"})
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("add by deck: got %d, want 201 (body=%s)", resp.StatusCode, body)
	}
	var added addBotResponse
	_ = json.NewDecoder(resp.Body).Decode(&added)
	resp.Body.Close()

	seat := added.Game.Players[1]
	if !seat.IsBot || seat.BotTier != "random" || seat.BotDeck != "test-mono-white" || !seat.DeckUploaded {
		t.Fatalf("bot seat: %+v", seat)
	}
	// …and the flags reached the game player, so PlayerView carries
	// them to the board without the client consulting the lobby.
	p := l.RoomOf(meta.ID).Game.PlayerByID(added.PlayerID)
	if p == nil || !p.IsBot || p.BotDeck != "test-mono-white" {
		t.Fatalf("game player: %+v", p)
	}

	// Unknown deck ID: 422, nothing seated.
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/seats/bot", joined.Token,
		addBotRequest{Tier: "random", Deck: "no-such-deck"})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("unknown deck: got %d, want 422", resp.StatusCode)
	}
	resp.Body.Close()

	// Both deck and source is a caller bug, not a silent precedence.
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/seats/bot", joined.Token,
		addBotRequest{Tier: "random", Deck: "test-mono-white", Format: "text", Source: "Commander:\n1 Test Commander\n"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("deck+source: got %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// Neither is too.
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/seats/bot", joined.Token, addBotRequest{Tier: "random"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("no deck: got %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	if got, _ := l.Get(meta.ID); len(got.Players) != 2 {
		t.Errorf("a rejected request seated something: %d players", len(got.Players))
	}
}

// A spectator's session carries this game's ID. Adding a bot mutates
// the table; watching does not earn it.
func TestSpectatorCannotSeatABot(t *testing.T) {
	srv, l, _, _ := newBotHTTPStack(t, buildMinimalDeckIndex(t), fakeDeckSource{})
	meta, _ := l.Create("FNM")
	_ = joinAs(t, srv, meta, "Alice")

	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/spectate", "",
		spectateRequest{InviteToken: meta.SpectatorInvite, Name: "Watcher"})
	var watcher sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&watcher)
	resp.Body.Close()
	if watcher.Token == "" {
		t.Fatal("spectate did not mint a session")
	}

	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/seats/bot", watcher.Token,
		addBotRequest{Tier: "random", Deck: "test-mono-white"})
	if resp.StatusCode != http.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("spectator add: got %d, want 403 (body=%s)", resp.StatusCode, body)
	}
	resp.Body.Close()
	if got, _ := l.Get(meta.ID); len(got.Players) != 1 {
		t.Errorf("a spectator seated a bot: %+v", got.Players)
	}
}

// A bot seat has to come back after a deploy, or the table resumes
// with an occupied chair nobody is sitting in and hangs the first
// time priority reaches it.
func TestBotRunnersRelaunchOnRestore(t *testing.T) {
	dir := t.TempDir()
	log := quietLogger()

	l, _ := newDurableLobby(t, dir)
	l.SetBotHost(newFakeBotHost())
	meta, err := l.Create("Survives a deploy")
	if err != nil {
		t.Fatal(err)
	}
	_, humanID, _ := l.Join(meta.ID, meta.InviteToken, "Alice")
	if _, err := l.SetDeck(meta.ID, humanID, "Alice's", botDeck(20)); err != nil {
		t.Fatal(err)
	}
	_, botID, err := l.AddBot(meta.ID, "Bot 1", "random", "test-mono-white", "Mono-white test", botDeck(20))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}

	// --- the deploy ------------------------------------------------
	l2, _ := newDurableLobby(t, dir)
	host2 := newFakeBotHost()
	l2.SetBotHost(host2)
	if n := l2.RestoreFromDisk(log); n != 1 {
		t.Fatalf("restored %d games, want 1", n)
	}

	back, err := l2.Get(meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(back.Players) != 2 {
		t.Fatalf("seats after restore: %+v", back.Players)
	}
	seat := back.Players[1]
	if !seat.IsBot || seat.BotTier != "random" || seat.BotDeck != "test-mono-white" {
		t.Errorf("bot seat metadata did not survive: %+v", seat)
	}
	// The engine snapshot carries it too, so PlayerView still says
	// "bot" to a client that reconnects after the deploy.
	p := l2.RoomOf(meta.ID).Game.PlayerByID(botID)
	if p == nil || !p.IsBot || p.BotTier != "random" || p.BotDeck != "test-mono-white" {
		t.Errorf("game player after restore: %+v", p)
	}
	// And the runner was relaunched.
	host2.mu.Lock()
	seats := host2.started[meta.ID]
	host2.mu.Unlock()
	if len(seats) != 1 || seats[0].PlayerID != botID || seats[0].Tier != "random" {
		t.Errorf("runners not relaunched after restore: %+v", seats)
	}
}

// A game restored in the LOBBY state has no runners to relaunch —
// Start will do it when someone presses the button.
func TestBotRunnersAreNotRelaunchedForAnUnstartedTable(t *testing.T) {
	dir := t.TempDir()
	log := quietLogger()
	l, _ := newDurableLobby(t, dir)
	meta, _ := l.Create("Still in the lobby")
	if _, _, err := l.AddBot(meta.ID, "Bot 1", "random", "", "d", botDeck(20)); err != nil {
		t.Fatal(err)
	}

	l2, _ := newDurableLobby(t, dir)
	host2 := newFakeBotHost()
	l2.SetBotHost(host2)
	if n := l2.RestoreFromDisk(log); n != 1 {
		t.Fatalf("restored %d games, want 1", n)
	}
	host2.mu.Lock()
	defer host2.mu.Unlock()
	if len(host2.started) != 0 {
		t.Errorf("started runners for an unstarted table: %+v", host2.started)
	}
}

// The placeholder deck this sub-PR ships must itself be a legal
// Commander deck — otherwise the one flow a human can actually take
// today ends in a 422, and the failure would only show up against a
// live Scryfall index. Resolved and validated here through the exact
// pipeline the endpoint uses.
func TestPlaceholderBotDeckIsLegal(t *testing.T) {
	idx := cards.NewIndex()
	idx.Put(cards.Card{
		ID:            uuid.New(),
		Name:          "Krenko, Mob Boss",
		TypeLine:      "Legendary Creature — Goblin Warrior",
		ColorIdentity: []string{"R"},
		Legalities:    map[string]string{"commander": "legal"},
	})
	idx.Put(cards.Card{
		ID:            uuid.New(),
		Name:          "Mountain",
		TypeLine:      "Basic Land — Mountain",
		ColorIdentity: []string{"R"},
		Legalities:    map[string]string{"commander": "legal"},
	})

	src := aiseat.PlaceholderDecks()
	for _, info := range src.List() {
		_, text, ok := src.Decklist(info.ID)
		if !ok {
			t.Fatalf("Decklist(%q) missing", info.ID)
		}
		entries, err := deck.ParseText(text)
		if err != nil {
			t.Fatalf("%s: ParseText: %v", info.ID, err)
		}
		list, err := deck.Resolve(idx, info.Name, entries)
		if err != nil {
			t.Fatalf("%s: Resolve: %v", info.ID, err)
		}
		if err := deck.Validate(list); err != nil {
			t.Fatalf("%s is not a legal Commander deck: %v", info.ID, err)
		}
		if got := len(list.Commanders) + len(list.Mainboard); got != 100 {
			t.Errorf("%s: %d cards, want 100", info.ID, got)
		}
	}
}
