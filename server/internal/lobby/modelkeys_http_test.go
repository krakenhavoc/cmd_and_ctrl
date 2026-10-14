package lobby

// Tests for ADR 0182's lobby half: /me/model-key, the own_key option of
// POST /games/{id}/seats/bot and GET /bot/options, and the seat's key
// binding through persistence.

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/modelkeys"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

const testModelKey = "sk-ant-FAKE-test-key-never-echo-9f8e7d6c"

type modelKeyStack struct {
	srv  *httptest.Server
	l    *Lobby
	host *fakeBotHost
	a    auth.Authenticator
	d    *db.DB
	keys *modelkeys.Service
	user uuid.UUID
}

// newModelKeyStack is a bot-capable server whose own model is OFF (the
// host offers random and heuristic only), with a database, a sealer and
// one signed-in person. on=false leaves the feature off.
func newModelKeyStack(t *testing.T, on bool) modelKeyStack {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	d := openTestDB(t, t.TempDir())
	s := modelKeyStack{d: d, user: uuid.New(), a: newTestAuth(), host: newFakeBotHost()}
	if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, 'Ian', 1, 1)`, s.user.String()); err != nil {
		t.Fatal(err)
	}
	if on {
		sealer, err := users.NewSealer(strings.Repeat("k", 40))
		if err != nil {
			t.Fatal(err)
		}
		s.keys = modelkeys.New(d, sealer, modelkeys.Options{Log: discardLogger()})
	}
	s.host.tiers = []string{"random", "heuristic"}
	s.l = NewLobbyWithStore(ws.NewRoomManager(discardLogger(), ""), NewSQLStore(d))
	s.l.SetBotHost(s.host)
	s.srv = newTestServerWithConfig(t, Config{
		Lobby: s.l, Auth: s.a, AdminToken: "shared-admin-token",
		Cards: buildMinimalDeckIndex(t), Bots: s.host, BotDecks: fakeDeckSource{},
		ModelKeys: s.keys,
	})
	return s
}

func (s modelKeyStack) session(t *testing.T, p auth.Principal) string {
	t.Helper()
	tok, _, err := s.a.Issue(context.Background(), p, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func (s modelKeyStack) identified(t *testing.T) string {
	return s.session(t, auth.Principal{Role: auth.RoleIdentified, UserID: s.user, Name: "Ian"})
}

func modelKeyCall(t *testing.T, srv *httptest.Server, method, path, tok string, body any) (int, string) {
	t.Helper()
	resp := doJSON(t, srv, method, path, tok, body)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(raw)
}

func noKeyIn(t *testing.T, where, body string) {
	t.Helper()
	if strings.Contains(body, testModelKey) || strings.Contains(body, "never-echo") {
		t.Errorf("%s returns the key: %s", where, body)
	}
}

func TestModelKeyRoutesAnswer503WhileTheFeatureIsOff(t *testing.T) {
	s := newModelKeyStack(t, false)
	tok := s.identified(t)
	for _, m := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		status, body := modelKeyCall(t, s.srv, m, "/me/model-key", tok, map[string]any{"provider": "anthropic"})
		if status != http.StatusServiceUnavailable || !strings.Contains(body, `"available":false`) {
			t.Errorf("%s: %d %s, want 503 available:false", m, status, body)
		}
	}
	if status, _ := modelKeyCall(t, s.srv, http.MethodPost, "/me/model-key/test", tok, nil); status != http.StatusServiceUnavailable {
		t.Errorf("test: %d, want 503", status)
	}
	// A guest is not a signed-in person, whatever the feature.
	guest := s.session(t, auth.Principal{Role: auth.RolePlayer, GameID: uuid.New(), PlayerID: uuid.New(), Name: "G"})
	if status, _ := modelKeyCall(t, s.srv, http.MethodGet, "/me/model-key", guest, nil); status != http.StatusForbidden {
		t.Errorf("guest: %d, want 403", status)
	}
}

func TestModelKeyRoundTripNeverReturnsTheKey(t *testing.T) {
	s := newModelKeyStack(t, true)
	tok := s.identified(t)

	status, body := modelKeyCall(t, s.srv, http.MethodGet, "/me/model-key", tok, nil)
	if status != http.StatusOK || !strings.Contains(body, `"key":null`) || !strings.Contains(body, `"custom"`) {
		t.Fatalf("GET before: %d %s", status, body)
	}

	put := map[string]any{"provider": "anthropic", "model": "claude-haiku-4-5", "key": testModelKey, "games_per_day": 3, "calls_per_game": 40}
	status, body = modelKeyCall(t, s.srv, http.MethodPut, "/me/model-key", tok, put)
	if status != http.StatusOK {
		t.Fatalf("PUT: %d %s", status, body)
	}
	noKeyIn(t, "PUT", body)
	var got modelKeyResponse
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatal(err)
	}
	if got.Key == nil || got.Key.Last4 != "7d6c" || got.Key.Provider != "anthropic" || got.Key.GamesPerDay != 3 {
		t.Errorf("PUT answered %+v", got.Key)
	}
	status, body = modelKeyCall(t, s.srv, http.MethodGet, "/me/model-key", tok, nil)
	if status != http.StatusOK || !strings.Contains(body, `"last4":"7d6c"`) {
		t.Errorf("GET after: %d %s", status, body)
	}
	noKeyIn(t, "GET", body)

	// Refusals say what to fix and never quote the body.
	bad := map[string]any{"provider": "custom", "base_url": "https://192.168.1.10/v1", "model": "m", "key": testModelKey, "games_per_day": 3, "calls_per_game": 40}
	status, body = modelKeyCall(t, s.srv, http.MethodPut, "/me/model-key", tok, bad)
	if status != http.StatusUnprocessableEntity || !strings.Contains(body, "private or local network") {
		t.Errorf("a private address: %d %s", status, body)
	}
	noKeyIn(t, "a refusal", body)
	status, body = modelKeyCall(t, s.srv, http.MethodPut, "/me/model-key", tok, map[string]any{"key": testModelKey, "surprise": 1})
	if status != http.StatusBadRequest {
		t.Errorf("an unknown field: %d", status)
	}
	noKeyIn(t, "a decode error", body)

	status, body = modelKeyCall(t, s.srv, http.MethodDelete, "/me/model-key", tok, nil)
	if status != http.StatusOK || !strings.Contains(body, `"key":null`) {
		t.Errorf("DELETE: %d %s", status, body)
	}
	if status, _ := modelKeyCall(t, s.srv, http.MethodPost, "/me/model-key/test", tok, nil); status != http.StatusNotFound {
		t.Errorf("test with no key: %d, want 404", status)
	}
}

// seatedOwner joins the table as a guest seat and returns a player
// session for that seat that also carries the person's user id, as a
// signed-in player's session does.
func (s modelKeyStack) seatedOwner(t *testing.T, meta GameMeta) (string, uuid.UUID) {
	t.Helper()
	joined := joinAs(t, s.srv, meta, "Ian")
	return s.session(t, auth.Principal{Role: auth.RolePlayer, UserID: s.user, GameID: meta.ID, PlayerID: joined.PlayerID, Name: "Ian"}), joined.PlayerID
}

func addOwnKeyBot(t *testing.T, s modelKeyStack, gameID uuid.UUID, tok string, body addBotRequest) (int, string) {
	t.Helper()
	resp := postJSON(t, s.srv, "/games/"+gameID.String()+"/seats/bot", tok, body)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(raw)
}

func TestABotRunsOnItsOwnersKeyOnAServerWithNoModel(t *testing.T) {
	s := newModelKeyStack(t, true)
	meta, _ := s.l.Create("FNM")
	tok, ownerPlayer := s.seatedOwner(t, meta)
	req := addBotRequest{Tier: "assisted", Deck: "test-mono-white", OwnKey: true}

	// No key yet.
	if status, body := addOwnKeyBot(t, s, meta.ID, tok, req); status != http.StatusUnprocessableEntity || !strings.Contains(body, "Settings › AI opponents") {
		t.Errorf("no key: %d %s", status, body)
	}
	if _, err := s.keys.Save(context.Background(), s.user, modelkeys.Input{
		Provider: "anthropic", Model: "claude-haiku-4-5", Key: testModelKey, GamesPerDay: 2, CallsPerGame: 40,
	}); err != nil {
		t.Fatal(err)
	}
	// Only a model tier can run on a key.
	if status, _ := addOwnKeyBot(t, s, meta.ID, tok, addBotRequest{Tier: "heuristic", Deck: "test-mono-white", OwnKey: true}); status != http.StatusUnprocessableEntity {
		t.Errorf("heuristic on a key: %d, want 422", status)
	}
	// A guest seat has no key to run on.
	guest := s.session(t, auth.Principal{Role: auth.RolePlayer, GameID: meta.ID, PlayerID: ownerPlayer, Name: "G"})
	if status, _ := addOwnKeyBot(t, s, meta.ID, guest, req); status != http.StatusForbidden {
		t.Errorf("a guest: %d, want 403", status)
	}
	// Without own_key the server's missing model refuses the tier.
	if status, _ := addOwnKeyBot(t, s, meta.ID, tok, addBotRequest{Tier: "assisted", Deck: "test-mono-white"}); status == http.StatusCreated {
		t.Error("assisted was seated on a server with no model and no own key")
	}
	// The picker offers the key.
	resp := doGet(t, s.srv, "/bot/options", tok)
	var opts botOptionsResponse
	_ = json.NewDecoder(resp.Body).Decode(&opts)
	resp.Body.Close()
	if opts.OwnKey == nil || opts.OwnKey.Model != "claude-haiku-4-5" || opts.OwnKey.ProviderLabel != "Anthropic" || !opts.OwnKey.GamesLeft {
		t.Errorf("own_key option = %+v", opts.OwnKey)
	}

	req.Model = "claude-sonnet-x"
	status, body := addOwnKeyBot(t, s, meta.ID, tok, req)
	if status != http.StatusCreated {
		t.Fatalf("own key: %d %s", status, body)
	}
	noKeyIn(t, "the add-bot answer", body)
	if strings.Contains(body, s.user.String()) {
		t.Error("the add-bot answer names the key owner's user id")
	}
	var added addBotResponse
	if err := json.Unmarshal([]byte(body), &added); err != nil {
		t.Fatal(err)
	}
	var seat SeatInfo
	for _, p := range added.Game.Players {
		if p.PlayerID == added.PlayerID && !p.BotOwnKey {
			t.Error("the answer does not mark the seat as on its owner's key")
		}
	}
	// The owner and model are server-side only; read the lobby's copy.
	live, _ := s.l.Get(meta.ID)
	for _, p := range live.Players {
		if p.PlayerID == added.PlayerID {
			seat = p
		}
	}
	if !seat.IsBot || seat.BotTier != "assisted" || !seat.BotOwnKey || seat.BotKeyOwner != s.user || seat.BotKeyModel != "claude-sonnet-x" {
		t.Errorf("seat = %+v", seat)
	}

	// The binding is on the row, so a restart keeps it.
	_, rows, err := NewSQLStore(s.d).LoadGame(context.Background(), meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range rows {
		if r.PlayerID == added.PlayerID {
			found = true
			if r.BotKeyUserID != s.user.String() || r.BotKeyModel != "claude-sonnet-x" {
				t.Errorf("row = %+v", r)
			}
		} else if r.BotKeyUserID != "" || r.BotKeyModel != "" {
			t.Errorf("a human seat carries a key binding: %+v", r)
		}
	}
	if !found {
		t.Fatal("no row for the bot")
	}

	// Start: the host is told whose key the seat runs on.
	if _, err := s.l.SetDeck(meta.ID, ownerPlayer, "Ian's", botDeck(20)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	s.host.mu.Lock()
	specs := s.host.started[meta.ID]
	s.host.mu.Unlock()
	if len(specs) != 1 || specs[0].KeyOwner != s.user || specs[0].KeyModel != "claude-sonnet-x" || specs[0].Tier != "assisted" {
		t.Errorf("started = %+v", specs)
	}
}

// A day's games used up refuses the seat at the picker, not at the
// seat's first move.
func TestAnOwnKeyBotIsRefusedWhenTheDaysGamesAreUsed(t *testing.T) {
	s := newModelKeyStack(t, true)
	if _, err := s.keys.Save(context.Background(), s.user, modelkeys.Input{
		Provider: "openai", Model: "gpt-x", Key: testModelKey, GamesPerDay: 1, CallsPerGame: 40,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.d.Exec(`UPDATE model_keys SET day = ?, day_games = 1 WHERE user_id = ?`,
		time.Now().UTC().Format(time.DateOnly), s.user.String()); err != nil {
		t.Fatal(err)
	}
	meta, _ := s.l.Create("FNM")
	tok, _ := s.seatedOwner(t, meta)
	status, body := addOwnKeyBot(t, s, meta.ID, tok, addBotRequest{Tier: "strong", Deck: "test-mono-white", OwnKey: true})
	if status != http.StatusUnprocessableEntity || !strings.Contains(body, "today") {
		t.Errorf("%d %s, want the daily-limit refusal", status, body)
	}
}

// A rematch carries an own-key bot only when its owner asks for it.
func TestARematchKeepsAnOwnKeyBotOnlyForItsOwner(t *testing.T) {
	l := newTestLobby(t)
	l.SetBotHost(newFakeBotHost())
	owner := uuid.New()
	meta, _ := l.Create("FNM")
	if _, _, err := l.AddBotOnKey(meta.ID, "Bot 1", "assisted", "", "Deck", botDeck(20), owner, ""); err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.AddBotOnKey(meta.ID, "Bot 2", "assisted", "", "Deck", botDeck(20), uuid.Nil, ""); err == nil {
		t.Error("an own-key bot with no owner was seated")
	}
	settings, err := l.tableSettings(meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	someoneElse := uuid.New()
	host := SeatInfo{PlayerID: uuid.New(), Name: "Ian", Seat: 0, DeckUploaded: true, UserID: owner.String()}
	mine := SeatInfo{PlayerID: uuid.New(), Name: "Mine", Seat: 1, IsBot: true, BotTier: "assisted", DeckUploaded: true,
		BotOwnKey: true, BotKeyOwner: owner, BotKeyModel: "m1"}
	theirs := SeatInfo{PlayerID: uuid.New(), Name: "Theirs", Seat: 2, IsBot: true, BotTier: "assisted", DeckUploaded: true,
		BotOwnKey: true, BotKeyOwner: someoneElse}
	plain := SeatInfo{PlayerID: uuid.New(), Name: "Plain", Seat: 3, IsBot: true, BotTier: "random", DeckUploaded: true}
	seats := []SeatInfo{host, mine, theirs, plain}
	decks := map[uuid.UUID][]game.Card{}
	for _, s := range seats {
		decks[s.PlayerID] = botDeck(20)
	}

	res, err := l.makeRematchTable("FNM again", seats, decks, settings, &host)
	if err != nil {
		t.Fatalf("makeRematchTable: %v", err)
	}
	names := map[string]SeatInfo{}
	for _, p := range res.Game.Players {
		names[p.Name] = p
	}
	if got, ok := names["Mine"]; !ok || got.BotKeyOwner != owner || got.BotKeyModel != "m1" || !got.BotOwnKey {
		t.Errorf("the host's own-key bot: %+v (present %v)", got, ok)
	}
	if _, ok := names["Theirs"]; ok {
		t.Error("a bot on someone else's key came along to a rematch they did not ask for")
	}
	if _, ok := names["Plain"]; !ok {
		t.Error("an ordinary bot was dropped")
	}
}
