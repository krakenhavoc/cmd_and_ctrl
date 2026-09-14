package lobby

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// fakeBotHost records what the lobby asks of it.
type fakeBotHost struct {
	mu      sync.Mutex
	started map[uuid.UUID][]aiseat.SeatSpec
	stopped []uuid.UUID
	tiers   []string
}

func newFakeBotHost() *fakeBotHost {
	return &fakeBotHost{started: map[uuid.UUID][]aiseat.SeatSpec{}, tiers: []string{"random"}}
}

func (f *fakeBotHost) Tiers() []string { return f.tiers }

func (f *fakeBotHost) StartBots(room *ws.Room, seats []aiseat.SeatSpec) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.started[room.Game.ID] = append([]aiseat.SeatSpec(nil), seats...)
}

func (f *fakeBotHost) StopBots(gameID uuid.UUID) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stopped = append(f.stopped, gameID)
}

// botDeck is a minimal legal-enough deck for AddBot: the lobby does
// not validate, the HTTP layer does.
func botDeck(n int) []game.Card {
	deck := []game.Card{game.NewCommander("Bot Commander", uuid.Nil)}
	for i := 0; i < n; i++ {
		c := game.NewCard(fmt.Sprintf("Mountain %d", i), uuid.Nil)
		c.TypeLine = "Basic Land — Mountain"
		deck = append(deck, c)
	}
	return deck
}

func TestAddBotSeatsADeckReadyBot(t *testing.T) {
	l := newTestLobby(t)
	host := newFakeBotHost()
	l.SetBotHost(host)
	meta, _ := l.Create("FNM")
	_, humanID, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}

	meta, botID, err := l.AddBot(meta.ID, "Bot 1", "random", "", "Mono Red", botDeck(20))
	if err != nil {
		t.Fatalf("AddBot: %v", err)
	}
	if len(meta.Players) != 2 {
		t.Fatalf("players: %d", len(meta.Players))
	}
	seat := meta.Players[1]
	if seat.PlayerID != botID || !seat.IsBot || seat.BotTier != "random" || !seat.DeckUploaded || seat.DeckName != "Mono Red" || seat.Seat != 1 {
		t.Errorf("bot seat: %+v", seat)
	}
	// The game-side player carries the flags too, so the view does.
	p := l.RoomOf(meta.ID).Game.PlayerByID(botID)
	if p == nil || !p.IsBot || p.BotTier != "random" || !p.DeckImported {
		t.Errorf("game player: %+v", p)
	}

	// Start: the human still needs a deck; then the host is told
	// about exactly the bot seat.
	if _, err := l.Start(meta.ID); err != ErrDeckNotUploaded {
		t.Fatalf("Start without the human's deck: %v", err)
	}
	if _, err := l.SetDeck(meta.ID, humanID, "Alice's", botDeck(20)); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	host.mu.Lock()
	seats := host.started[meta.ID]
	host.mu.Unlock()
	if len(seats) != 1 || seats[0].PlayerID != botID || seats[0].Tier != "random" {
		t.Errorf("host started: %+v", seats)
	}
	// Delete stops them.
	if err := l.Delete(meta.ID); err != nil {
		t.Fatal(err)
	}
	host.mu.Lock()
	defer host.mu.Unlock()
	if len(host.stopped) != 1 || host.stopped[0] != meta.ID {
		t.Errorf("host stopped: %v", host.stopped)
	}
}

func TestAddBotRespectsTableLimitsAndState(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("FNM")
	for i := 0; i < game.MaxPlayers; i++ {
		if _, _, err := l.AddBot(meta.ID, fmt.Sprintf("Bot %d", i), "random", "", "d", botDeck(5)); err != nil {
			t.Fatalf("AddBot %d: %v", i, err)
		}
	}
	if _, _, err := l.AddBot(meta.ID, "Bot 5", "random", "", "d", botDeck(5)); err != ErrGameFull {
		t.Errorf("fifth seat: %v, want ErrGameFull", err)
	}
	if _, _, err := l.AddBot(meta.ID, "", "random", "", "d", botDeck(5)); err != ErrEmptyName {
		t.Errorf("empty name: %v, want ErrEmptyName", err)
	}
	if _, _, err := l.AddBot(meta.ID, "Bot", "random", "", "d", nil); err != ErrDeckNotUploaded {
		t.Errorf("no deck: %v, want ErrDeckNotUploaded", err)
	}
	// An all-bot table starts (the fuzz harness case).
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if _, _, err := l.AddBot(meta.ID, "Late", "random", "", "d", botDeck(5)); err != ErrGameStarted {
		t.Errorf("add after start: %v, want ErrGameStarted", err)
	}
}

func TestRemoveBotClosesTheSeatGap(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("FNM")
	_, humanID, _ := l.Join(meta.ID, meta.InviteToken, "Alice")
	_, bot1, _ := l.AddBot(meta.ID, "Bot 1", "random", "", "d", botDeck(5))
	_, bot2, _ := l.AddBot(meta.ID, "Bot 2", "random", "", "d", botDeck(5))

	if _, err := l.RemoveBot(meta.ID, humanID); err != ErrNotABot {
		t.Errorf("removing a human: %v, want ErrNotABot", err)
	}
	if _, err := l.RemoveBot(meta.ID, uuid.New()); err != ErrPlayerNotInGame {
		t.Errorf("removing a stranger: %v, want ErrPlayerNotInGame", err)
	}
	meta, err := l.RemoveBot(meta.ID, bot1)
	if err != nil {
		t.Fatalf("RemoveBot: %v", err)
	}
	if len(meta.Players) != 2 || meta.Players[0].PlayerID != humanID || meta.Players[1].PlayerID != bot2 {
		t.Fatalf("players after remove: %+v", meta.Players)
	}
	if meta.Players[1].Seat != 1 {
		t.Errorf("seat gap not closed in meta: %+v", meta.Players[1])
	}
	g := l.RoomOf(meta.ID).Game
	if len(g.Seats) != 2 || g.Seats[1].ID != bot2 || g.Seats[1].Seat != 1 {
		t.Errorf("seat gap not closed in game: %+v", g.Seats)
	}
	// A third bot takes seat 2, not 3.
	meta, _, err = l.AddBot(meta.ID, "Bot 3", "random", "", "d", botDeck(5))
	if err != nil {
		t.Fatal(err)
	}
	if meta.Players[2].Seat != 2 {
		t.Errorf("new seat: %+v", meta.Players[2])
	}
}

// --- HTTP ---------------------------------------------------------

func TestBotRoutesRoundTrip(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	log := discardLogger()
	mgr := ws.NewRoomManager(log, "")
	l := NewLobby(mgr)
	host := newFakeBotHost()
	l.SetBotHost(host)
	srv := newTestServerWithConfig(t, Config{Lobby: l, Auth: newTestAuth(), AdminToken: "shared-admin-token", Cards: idx, Bots: host})

	meta, _ := l.Create("FNM")
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	var joined sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&joined)
	resp.Body.Close()

	source := "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n"
	// Unknown tier → 422.
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/seats/bot", joined.Token,
		addBotRequest{Tier: "galaxy-brain", Format: "text", Source: source})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unknown tier: got %d (body=%s)", resp.StatusCode, body)
	}
	resp.Body.Close()

	// A seated player adds a bot with a decklist; default name.
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/seats/bot", joined.Token,
		addBotRequest{Tier: "random", Format: "text", Source: source})
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("add bot: got %d, want 201 (body=%s)", resp.StatusCode, body)
	}
	var added addBotResponse
	_ = json.NewDecoder(resp.Body).Decode(&added)
	resp.Body.Close()
	if added.PlayerID == uuid.Nil || len(added.Game.Players) != 2 {
		t.Fatalf("add bot response: %+v", added)
	}
	bot := added.Game.Players[1]
	if !bot.IsBot || bot.BotTier != "random" || bot.Name != "Bot 1" || !bot.DeckUploaded {
		t.Errorf("bot seat: %+v", bot)
	}
	// A vanilla commander and basic lands need no catalog entry, so
	// there is nothing to disclose here. See
	// TestAddBotDisclosesUnimplementedCards for the other half.
	if len(added.Unimplemented) != 0 {
		t.Errorf("a vanilla deck reported unimplemented cards: %v", added.Unimplemented)
	}

	// A bad deck is a 422 with violations, and seats nothing.
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/seats/bot", joined.Token,
		addBotRequest{Tier: "random", Format: "text", Source: "Commander:\n1 Test Commander\nMainboard:\n3 Plains\n"})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("short deck: got %d, want 422 (body=%s)", resp.StatusCode, body)
	}
	resp.Body.Close()
	if got, _ := l.Get(meta.ID); len(got.Players) != 2 {
		t.Errorf("a rejected deck seated a bot: %d players", len(got.Players))
	}

	// A stranger's session cannot touch the table.
	other, _ := l.Create("Other")
	resp = postJSON(t, srv, "/games/"+other.ID.String()+"/join", "",
		joinRequest{InviteToken: other.InviteToken, Name: "Mallory"})
	var mallory sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&mallory)
	resp.Body.Close()
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/seats/bot", mallory.Token,
		addBotRequest{Tier: "random", Format: "text", Source: source})
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("stranger add: got %d, want 403", resp.StatusCode)
	}
	resp.Body.Close()

	// Remove it.
	resp = doDelete(t, srv, "/games/"+meta.ID.String()+"/seats/bot/"+added.PlayerID.String(), joined.Token)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("remove bot: got %d (body=%s)", resp.StatusCode, body)
	}
	var after GameMeta
	_ = json.NewDecoder(resp.Body).Decode(&after)
	resp.Body.Close()
	if len(after.Players) != 1 {
		t.Errorf("players after remove: %+v", after.Players)
	}
	// Removing the human is refused.
	resp = doDelete(t, srv, "/games/"+meta.ID.String()+"/seats/bot/"+joined.PlayerID.String(), joined.Token)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("remove human: got %d, want 422", resp.StatusCode)
	}
	resp.Body.Close()
}

// TestAddBotDisclosesUnimplementedCards covers issue #89's
// "catalog-gap tolerance" from the seating end. The curated decks are
// build-tested to be fully covered, so this can only happen through
// the raw-decklist escape hatch — which is exactly where it matters,
// because that is the path by which a card the engine cannot execute
// reaches a bot seat. POST /games/{id}/decks has disclosed this since
// the five mid-game-surprise bug reports of 2026-09-10; the bot route
// was installing decks without it.
func TestAddBotDisclosesUnimplementedCards(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	// A card whose printed text needs a catalog Spec. These tests wire
	// no effect catalog, so it resolves to nothing and the engine will
	// not carry it out.
	idx.Put(cards.Card{
		ID:            uuid.New(),
		Name:          "Test Wrath",
		TypeLine:      "Sorcery",
		OracleText:    "Destroy all creatures.",
		ColorIdentity: []string{"W"},
		Legalities:    map[string]string{"commander": "legal"},
	})
	log := discardLogger()
	l := NewLobby(ws.NewRoomManager(log, ""))
	host := newFakeBotHost()
	l.SetBotHost(host)
	srv := newTestServerWithConfig(t, Config{Lobby: l, Auth: newTestAuth(), AdminToken: "shared-admin-token", Cards: idx, Bots: host})

	meta, _ := l.Create("FNM")
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	var joined sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&joined)
	resp.Body.Close()

	source := "Commander:\n1 Test Commander\nMainboard:\n1 Test Wrath\n98 Plains\n"
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/seats/bot", joined.Token,
		addBotRequest{Tier: "random", Format: "text", Source: source})
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("add bot: got %d, want 201 (body=%s)", resp.StatusCode, body)
	}
	var added addBotResponse
	_ = json.NewDecoder(resp.Body).Decode(&added)
	resp.Body.Close()

	if len(added.Unimplemented) != 1 || added.Unimplemented[0] != "Test Wrath" {
		t.Errorf("unimplemented = %v, want [Test Wrath] — a bot seated with a card the engine cannot execute must say so", added.Unimplemented)
	}
	// The seat is still taken: a catalog gap is a disclosure, not a
	// refusal. Refusing would hold a bot's deck to a stricter standard
	// than a human's.
	if got, _ := l.Get(meta.ID); len(got.Players) != 2 {
		t.Errorf("the disclosure blocked the seat: %d players", len(got.Players))
	}
}

func TestBotRoutesDisabledWithoutHost(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	srv, l, _, _ := newTestHTTPStackWithCards(t, idx)
	meta, _ := l.Create("FNM")
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	var joined sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&joined)
	resp.Body.Close()
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/seats/bot", joined.Token,
		addBotRequest{Tier: "random", Format: "text", Source: "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n"})
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("no host: got %d, want 503", resp.StatusCode)
	}
	resp.Body.Close()
}

// --- helpers local to the bot tests ---------------------------------

func discardLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func newTestAuth() auth.Authenticator { return auth.NewMemoryAuthenticator() }

// newTestServerWithConfig stands up the lobby handler on the given
// config; the bot tests need Config.Bots, which the shared stacks
// don't set.
func newTestServerWithConfig(t *testing.T, cfg Config) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle("/", Handler(cfg))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func doDelete(t *testing.T, srv *httptest.Server, path, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, srv.URL+path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	return resp
}
