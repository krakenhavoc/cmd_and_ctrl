package lobby

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decks"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// practice_test.go covers the tutorial's practice table (ADR 0076
// §2.2, #1078): the lobby half in practice.go and the routes in
// practice_http.go.

// fakeEvictor records the games the lobby evicted on its own.
type fakeEvictor struct {
	mu      sync.Mutex
	evicted []uuid.UUID
}

func (f *fakeEvictor) EvictGame(id uuid.UUID) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.evicted = append(f.evicted, id)
	return 0
}

func (f *fakeEvictor) has(id uuid.UUID) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, e := range f.evicted {
		if e == id {
			return true
		}
	}
	return false
}

func (f *fakeBotHost) stoppedGame(id uuid.UUID) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range f.stopped {
		if s == id {
			return true
		}
	}
	return false
}

func (f *fakeBotHost) startedSeats(id uuid.UUID) []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []string
	for _, s := range f.started[id] {
		out = append(out, s.Tier+"/"+s.Deck)
	}
	return out
}

func newPracticeLobby(t *testing.T) (*Lobby, *fakeBotHost, *fakeEvictor) {
	t.Helper()
	l := NewLobby(ws.NewRoomManager(discardLogger(), ""))
	host := newFakeBotHost()
	l.SetBotHost(host)
	ev := &fakeEvictor{}
	l.SetEvictor(ev)
	return l, host, ev
}

func practiceSeats(owner string) (PracticeHuman, PracticeBot) {
	return PracticeHuman{Name: "Alice", Owner: owner, DeckName: "First Steps", Cards: botDeck(20)},
		PracticeBot{Name: "Practice Bot", Tier: "random", DeckID: decks.TutorialBotDeckID, DeckName: "Practice Partner", Cards: botDeck(20)}
}

func TestCreatePracticeSeatsThePlayerFirstAndStartsTheBot(t *testing.T) {
	l, host, _ := newPracticeLobby(t)
	human, bot := practiceSeats("user:a")
	meta, playerID, err := l.CreatePractice(human, bot)
	if err != nil {
		t.Fatalf("CreatePractice: %v", err)
	}
	if !meta.Practice || meta.State != string(game.StateActive) {
		t.Errorf("meta: practice=%v state=%q, want a started practice table", meta.Practice, meta.State)
	}
	if len(meta.Players) != 2 {
		t.Fatalf("seats: %+v", meta.Players)
	}
	me, them := meta.Players[0], meta.Players[1]
	if me.PlayerID != playerID || me.Seat != 0 || me.IsBot || !me.DeckUploaded || !me.IsHost {
		t.Errorf("human seat: %+v", me)
	}
	if !them.IsBot || them.BotTier != "random" || them.BotDeck != decks.TutorialBotDeckID || them.Seat != 1 {
		t.Errorf("bot seat: %+v", them)
	}
	if got := host.startedSeats(meta.ID); len(got) != 1 || got[0] != "random/"+decks.TutorialBotDeckID {
		t.Errorf("bot runners started: %v", got)
	}

	// The human takes turn one: the tutorial's steps are ordered by
	// the player's first turn.
	g, err := l.LookupGame(meta.ID)
	if err != nil {
		t.Fatalf("LookupGame: %v", err)
	}
	if g.StartingSeat != 0 || g.Turn.ActiveSeat != 0 {
		t.Errorf("starting seat %d, active seat %d; the player goes first", g.StartingSeat, g.Turn.ActiveSeat)
	}
}

func TestPracticeTableIsUnlistedAndUnpersisted(t *testing.T) {
	l, _, _ := newPracticeLobby(t)
	real, err := l.Create("A real table")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	human, bot := practiceSeats("user:a")
	meta, _, err := l.CreatePractice(human, bot)
	if err != nil {
		t.Fatalf("CreatePractice: %v", err)
	}

	for _, m := range l.List() {
		if m.ID == meta.ID {
			t.Error("GET /games lists the practice table")
		}
	}
	if got := l.List(); len(got) != 1 || got[0].ID != real.ID {
		t.Errorf("listing: %+v", got)
	}
	if _, err := l.Get(meta.ID); err != nil {
		t.Errorf("Get on the practice table: %v", err)
	}

	ctx := context.Background()
	if _, _, err := l.store.LoadGame(ctx, meta.ID); !errors.Is(err, ErrStoreNotFound) {
		t.Errorf("the practice table reached the store (err=%v)", err)
	}
	ids, _ := l.store.GameIDs(ctx)
	for _, id := range ids {
		if id == meta.ID {
			t.Error("the store has a games row for the practice table")
		}
	}
	if l.RoomOf(meta.ID).ReplayPath() != "" {
		t.Error("the practice table's room writes a replay; it must leave nothing on disk")
	}
}

func TestCreatePracticeReplacesTheOwnersLastTable(t *testing.T) {
	l, host, ev := newPracticeLobby(t)
	human, bot := practiceSeats("user:a")
	first, _, err := l.CreatePractice(human, bot)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	otherHuman, otherBot := practiceSeats("user:b")
	other, _, err := l.CreatePractice(otherHuman, otherBot)
	if err != nil {
		t.Fatalf("other owner: %v", err)
	}
	second, _, err := l.CreatePractice(human, bot)
	if err != nil {
		t.Fatalf("second: %v", err)
	}

	if _, err := l.Get(first.ID); !errors.Is(err, ErrGameNotFound) {
		t.Errorf("the owner's first table survived their second (err=%v)", err)
	}
	if !host.stoppedGame(first.ID) || !ev.has(first.ID) {
		t.Errorf("the replaced table's runner was not stopped or its sockets not closed")
	}
	for _, id := range []uuid.UUID{second.ID, other.ID} {
		if _, err := l.Get(id); err != nil {
			t.Errorf("table %s: %v", id, err)
		}
	}
}

func TestCreatePracticeHasACeiling(t *testing.T) {
	l, _, _ := newPracticeLobby(t)
	l.SetPracticeLimits(PracticeLimits{MaxTables: 2})
	for _, owner := range []string{"user:a", "user:b"} {
		h, b := practiceSeats(owner)
		if _, _, err := l.CreatePractice(h, b); err != nil {
			t.Fatalf("%s: %v", owner, err)
		}
	}
	h, b := practiceSeats("user:c")
	if _, _, err := l.CreatePractice(h, b); !errors.Is(err, ErrPracticeTablesFull) {
		t.Fatalf("third owner: got %v, want ErrPracticeTablesFull", err)
	}
	// An owner who already has one replaces it, so is not refused.
	h, b = practiceSeats("user:a")
	if _, _, err := l.CreatePractice(h, b); err != nil {
		t.Errorf("an owner replacing their own table was refused: %v", err)
	}
	// Real tables do not count.
	if _, err := l.Create("A real table"); err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestLeavePractice(t *testing.T) {
	l, host, ev := newPracticeLobby(t)
	human, bot := practiceSeats("user:a")
	meta, playerID, err := l.CreatePractice(human, bot)
	if err != nil {
		t.Fatalf("CreatePractice: %v", err)
	}
	real, _ := l.Create("A real table")

	if err := l.LeavePractice(meta.ID, meta.Players[1].PlayerID); !errors.Is(err, ErrPlayerNotInGame) {
		t.Errorf("the bot's seat left: %v", err)
	}
	if err := l.LeavePractice(meta.ID, uuid.New()); !errors.Is(err, ErrPlayerNotInGame) {
		t.Errorf("a stranger left: %v", err)
	}
	if err := l.LeavePractice(real.ID, playerID); !errors.Is(err, ErrNotPracticeTable) {
		t.Errorf("leaving a real table: got %v, want ErrNotPracticeTable", err)
	}
	if _, err := l.Get(real.ID); err != nil {
		t.Errorf("the real table is gone: %v", err)
	}

	if err := l.LeavePractice(meta.ID, playerID); err != nil {
		t.Fatalf("LeavePractice: %v", err)
	}
	if _, err := l.Get(meta.ID); !errors.Is(err, ErrGameNotFound) {
		t.Errorf("the table survived its player leaving (err=%v)", err)
	}
	if !host.stoppedGame(meta.ID) || !ev.has(meta.ID) {
		t.Error("leaving did not stop the bot runner or close the sockets")
	}
	if err := l.LeavePractice(meta.ID, playerID); !errors.Is(err, ErrGameNotFound) {
		t.Errorf("leaving twice: got %v, want ErrGameNotFound", err)
	}
}

// TestIdlePracticeTableIsReaped is the closed-tab path the server can
// see: nothing commits, so the table goes.
func TestIdlePracticeTableIsReaped(t *testing.T) {
	l, host, ev := newPracticeLobby(t)
	l.SetPracticeLimits(PracticeLimits{Idle: 50 * time.Millisecond})
	human, bot := practiceSeats("user:a")
	meta, _, err := l.CreatePractice(human, bot)
	if err != nil {
		t.Fatalf("CreatePractice: %v", err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := l.Get(meta.ID); errors.Is(err, ErrGameNotFound) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("an idle practice table was never reaped")
		}
		time.Sleep(10 * time.Millisecond)
	}
	// The reap's stop-and-evict runs just after the registry drops
	// the table; give it the same budget.
	for !host.stoppedGame(meta.ID) || !ev.has(meta.ID) {
		if time.Now().After(deadline) {
			t.Fatal("the reap did not stop the bot runner and close the sockets")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// --- routes ---

// tutorialDeckIndex is a card index holding exactly the tutorial pair's
// cards, so the practice route can load both decks offline. The type
// lines are stand-ins; deck.Validate needs a legendary creature in the
// command zone and basic lands among the rest.
func tutorialDeckIndex(t *testing.T) *cards.Index {
	t.Helper()
	idx := cards.NewIndex()
	for _, d := range []decks.Deck{decks.TutorialPlayer(), decks.TutorialBot()} {
		for i, c := range d.Cards() {
			card := cards.Card{
				ID:         uuid.New(),
				Name:       c.Name,
				TypeLine:   "Creature — Test",
				Legalities: map[string]string{"commander": "legal"},
			}
			for _, r := range c.Identity {
				card.ColorIdentity = append(card.ColorIdentity, string(r))
			}
			switch {
			case i == 0:
				card.TypeLine = "Legendary Creature — Test"
			case c.Basic:
				card.TypeLine = "Basic Land — " + c.Name
				card.ColorIdentity = []string{"G"}
			}
			if c.OracleID != "" {
				card.OracleID = uuid.MustParse(c.OracleID)
			}
			idx.Put(card)
		}
	}
	return idx
}

func newPracticeServer(t *testing.T) (srvURL func(string) string, do func(method, path, token, contentType, body string) *http.Response, l *Lobby, host *fakeBotHost) {
	t.Helper()
	// The round trip makes more unauthenticated calls than the join
	// bucket's burst allows; the limiter is not what is under test.
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	l, host, _ = newPracticeLobby(t)
	srv := newTestServerWithConfig(t, Config{
		Lobby: l, Auth: newTestAuth(), AdminToken: "shared-admin-token",
		Cards: tutorialDeckIndex(t), Bots: host,
	})
	srvURL = func(p string) string { return srv.URL + p }
	do = func(method, path, token, contentType, body string) *http.Response {
		t.Helper()
		req, err := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
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
	return srvURL, do, l, host
}

func sessionCookie(resp *http.Response) (*http.Cookie, bool) {
	for _, c := range resp.Cookies() {
		if c.Name == auth.SessionCookie {
			return c, true
		}
	}
	return nil, false
}

func TestPracticeRoutesRoundTrip(t *testing.T) {
	_, do, l, host := newPracticeServer(t)

	// A guest seated at a real table opens the tutorial.
	real, _ := l.Create("FNM")
	resp := do(http.MethodPost, "/games/"+real.ID.String()+"/join", "", "application/json",
		`{"invite_token":"`+real.InviteToken+`","name":"Alice"}`)
	var guest sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&guest)
	resp.Body.Close()
	if guest.Token == "" {
		t.Fatalf("join: status %d", resp.StatusCode)
	}

	resp = do(http.MethodPost, "/games/practice", guest.Token, "application/json", "")
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("create: got %d, want 201 (body=%s)", resp.StatusCode, body)
	}
	var practice sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&practice)
	resp.Body.Close()
	if practice.Game == nil || !practice.Game.Practice || practice.Game.ID == real.ID {
		t.Fatalf("create response: %+v", practice)
	}
	if c, ok := sessionCookie(resp); !ok || c.Value != practice.Token {
		t.Error("create did not set the session cookie to the practice seat")
	}
	p := practice.Principal
	if p.Role != auth.RolePlayer || p.GameID != practice.Game.ID || p.PlayerID != practice.PlayerID || p.Name != "Alice" {
		t.Errorf("practice session: %+v", p)
	}
	seats := practice.Game.Players
	if len(seats) != 2 || seats[0].PlayerID != practice.PlayerID || seats[0].DeckName != "First Steps" ||
		!seats[1].IsBot || seats[1].DeckName != "Practice Partner" {
		t.Errorf("seats: %+v", seats)
	}
	if got := host.startedSeats(practice.Game.ID); len(got) != 1 {
		t.Errorf("bot runner not started: %v", got)
	}

	// Not listed, to the player or anyone else.
	resp = do(http.MethodGet, "/games", practice.Token, "", "")
	var listed []GameMeta
	_ = json.NewDecoder(resp.Body).Decode(&listed)
	resp.Body.Close()
	for _, m := range listed {
		if m.ID == practice.Game.ID {
			t.Error("GET /games lists the practice table")
		}
	}

	// Leave refuses a body a form could send.
	leavePath := "/games/" + practice.Game.ID.String() + "/practice/leave"
	resp = do(http.MethodPost, leavePath, "", "text/plain",
		`{"practice_token":"`+practice.Token+`","restore_token":"`+guest.Token+`"}`)
	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("text/plain leave: got %d, want 415", resp.StatusCode)
	}
	resp.Body.Close()
	if _, err := l.Get(practice.Game.ID); err != nil {
		t.Fatalf("a refused leave closed the table: %v", err)
	}

	// Leave: the table goes, and the cookie is the guest's again.
	resp = do(http.MethodPost, leavePath, "", "application/json",
		`{"practice_token":"`+practice.Token+`","restore_token":"`+guest.Token+`"}`)
	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("leave: got %d, want 204 (body=%s)", resp.StatusCode, body)
	}
	resp.Body.Close()
	if c, ok := sessionCookie(resp); !ok || c.Value != guest.Token {
		t.Errorf("leave did not restore the guest's cookie (got %+v)", c)
	}
	if _, err := l.Get(practice.Game.ID); !errors.Is(err, ErrGameNotFound) {
		t.Errorf("the practice table survived leave (err=%v)", err)
	}
	if !host.stoppedGame(practice.Game.ID) {
		t.Error("leave did not stop the bot runner")
	}

	// Again — the next page load does exactly this — is still a 204,
	// and still restores the cookie.
	resp = do(http.MethodPost, leavePath, "", "application/json",
		`{"practice_token":"`+practice.Token+`","restore_token":"`+guest.Token+`"}`)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("second leave: got %d, want 204", resp.StatusCode)
	}
	if c, ok := sessionCookie(resp); !ok || c.Value != guest.Token {
		t.Errorf("second leave did not restore the cookie (got %+v)", c)
	}
	resp.Body.Close()

	// A dead restore token clears the cookie instead of keeping the
	// practice seat's.
	resp = do(http.MethodPost, leavePath, "", "application/json", `{"practice_token":"","restore_token":"not-a-token"}`)
	if c, ok := sessionCookie(resp); !ok || c.Value != "" || c.MaxAge >= 0 {
		t.Errorf("an invalid restore token did not clear the cookie (got %+v)", c)
	}
	resp.Body.Close()

	// The real table's own seat cannot use the route to delete it.
	resp = do(http.MethodPost, "/games/"+real.ID.String()+"/practice/leave", "", "application/json",
		`{"practice_token":"`+guest.Token+`"}`)
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("leave on a real table: got %d, want 409", resp.StatusCode)
	}
	resp.Body.Close()
	if _, err := l.Get(real.ID); err != nil {
		t.Errorf("the real table is gone: %v", err)
	}
}

func TestPracticeRouteRefusesWithoutABotHostOrCards(t *testing.T) {
	l := NewLobby(ws.NewRoomManager(discardLogger(), ""))
	a := newTestAuth()
	tok, _, err := a.Issue(context.Background(), auth.Principal{Role: auth.RoleAdmin, AdminID: uuid.New(), Name: "admin"}, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	for name, cfg := range map[string]Config{
		"no bot host": {Lobby: l, Auth: a, Cards: tutorialDeckIndex(t)},
		"no cards":    {Lobby: l, Auth: a, Bots: newFakeBotHost()},
	} {
		srv := newTestServerWithConfig(t, cfg)
		resp := postJSON(t, srv, "/games/practice", tok, struct{}{})
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("%s: got %d, want 503", name, resp.StatusCode)
		}
		resp.Body.Close()
	}
	srv := newTestServerWithConfig(t, Config{Lobby: l, Auth: a, Cards: tutorialDeckIndex(t), Bots: newFakeBotHost()})
	resp := postJSON(t, srv, "/games/practice", "", struct{}{})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no session: got %d, want 401", resp.StatusCode)
	}
	resp.Body.Close()
}

// TestPracticeOwner pins who a practice table belongs to, which is what
// decides whether opening a second one replaces the first.
func TestPracticeOwner(t *testing.T) {
	l, _, _ := newPracticeLobby(t)
	c := Config{Lobby: l}
	user, seat, gid := uuid.New(), uuid.New(), uuid.New()
	cases := []struct {
		name string
		p    auth.Principal
		want string
	}{
		{"signed-in user", auth.Principal{Role: auth.RoleIdentified, UserID: user, DiscordID: "1"}, "user:" + user.String()},
		{"discord, no database", auth.Principal{Role: auth.RoleIdentified, DiscordID: "42"}, "discord:42"},
		{"admin", auth.Principal{Role: auth.RoleAdmin, AdminID: uuid.New()}, "admin"},
		{"guest seat", auth.Principal{Role: auth.RolePlayer, GameID: gid, PlayerID: seat}, "seat:" + seat.String()},
		{"spectator", auth.Principal{Role: auth.RoleSpectator, GameID: gid}, ""},
	}
	for _, tc := range cases {
		if got := practiceOwner(c, tc.p); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}

	// A practice seat belongs to whoever opened its table.
	h, b := practiceSeats("seat:" + seat.String())
	meta, playerID, err := l.CreatePractice(h, b)
	if err != nil {
		t.Fatalf("CreatePractice: %v", err)
	}
	p := auth.Principal{Role: auth.RolePlayer, GameID: meta.ID, PlayerID: playerID}
	if got := practiceOwner(c, p); got != "seat:"+seat.String() {
		t.Errorf("practice seat: got %q, want the opener's key", got)
	}
}
