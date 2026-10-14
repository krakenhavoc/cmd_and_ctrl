package lobby

// Tests for ADR 0162's lobby half: the routes GET and PUT
// /me/notifications, the fill and start events, and the tables the
// notifier is shown (notify.go).

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/notify"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

func newNotificationsStack(t *testing.T, on bool) (userStack, *notify.SQLStore) {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	var store *notify.SQLStore
	s := newUserStackWithDB(t, nil, func(c *Config, d *db.DB) {
		store = notify.NewSQLStore(d)
		if on {
			c.DiscordBot = discord.Bot{Token: "bot-token"}
			c.InviteBaseURL = "https://cmd.example"
			c.Notifications = store
		}
	})
	return s, store
}

func notificationsCall(t *testing.T, s userStack, method, tok string, body any) (int, notificationsResponse) {
	t.Helper()
	resp := doJSON(t, s.srv, method, "/me/notifications", tok, body)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out notificationsResponse
	_ = json.Unmarshal(raw, &out)
	return resp.StatusCode, out
}

func TestNotificationsAreOffByDefault(t *testing.T) {
	s, _ := newNotificationsStack(t, true)
	tok, _ := signedInSettingsUser(t, s)
	status, got := notificationsCall(t, s, http.MethodGet, tok, nil)
	if status != http.StatusOK {
		t.Fatalf("GET: %d", status)
	}
	want := notificationsResponse{Available: true}
	if got != want {
		t.Errorf("GET for someone who never saved = %+v, want %+v", got, want)
	}
}

func TestNotificationsPutChangesOnlyTheNamedSwitch(t *testing.T) {
	s, _ := newNotificationsStack(t, true)
	tok, _ := signedInSettingsUser(t, s)
	if status, got := notificationsCall(t, s, http.MethodPut, tok, map[string]bool{"turn": true, "table_filled": true}); status != http.StatusOK || !got.Turn || !got.TableFilled {
		t.Fatalf("PUT: %d %+v", status, got)
	}
	status, got := notificationsCall(t, s, http.MethodPut, tok, map[string]bool{"table_filled": false})
	if status != http.StatusOK || !got.Turn || got.TableReady || got.TableFilled {
		t.Errorf("second PUT: %d %+v, want turn alone on", status, got)
	}
	if _, again := notificationsCall(t, s, http.MethodGet, tok, nil); again != got {
		t.Errorf("GET = %+v, PUT answered %+v", again, got)
	}
	for _, bad := range []any{map[string]bool{}, map[string]any{"turn": "yes"}, map[string]bool{"chat": true}} {
		if status, _ := notificationsCall(t, s, http.MethodPut, tok, bad); status != http.StatusBadRequest {
			t.Errorf("PUT %v: %d, want 400", bad, status)
		}
	}
}

func TestNotificationsShowAndClearARefusal(t *testing.T) {
	s, store := newNotificationsStack(t, true)
	tok, user := signedInSettingsUser(t, s)
	notificationsCall(t, s, http.MethodPut, tok, map[string]bool{"turn": true})
	if err := store.MarkFailed(context.Background(), user, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, got := notificationsCall(t, s, http.MethodGet, tok, nil); !got.DMFailed {
		t.Fatalf("after a refusal GET = %+v, want dm_failed", got)
	}
	notificationsCall(t, s, http.MethodPut, tok, map[string]bool{"turn": false})
	if _, got := notificationsCall(t, s, http.MethodPut, tok, map[string]bool{"turn": true}); got.DMFailed {
		t.Errorf("turning the switch off and on again left dm_failed: %+v", got)
	}
}

// No bot token: the feature is off, both routes answer 503 with
// "available": false, and nothing is written.
func TestNotificationsWithNoBotTokenAre503(t *testing.T) {
	s, _ := newNotificationsStack(t, false)
	tok, _ := signedInSettingsUser(t, s)
	for _, method := range []string{http.MethodGet, http.MethodPut} {
		status, got := notificationsCall(t, s, method, tok, map[string]bool{"turn": true})
		if status != http.StatusServiceUnavailable || got.Available {
			t.Errorf("%s with no bot token: %d %+v, want 503 and available false", method, status, got)
		}
	}
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM notification_prefs`).Scan(&n); err != nil || n != 0 {
		t.Errorf("notification_prefs rows = %d (%v), want 0", n, err)
	}
}

func TestNotificationsNeedASignedInPerson(t *testing.T) {
	s, _ := newNotificationsStack(t, true)
	meta, err := s.lobby.Create("Pod")
	if err != nil {
		t.Fatal(err)
	}
	_, playerID, err := s.lobby.Join(meta.ID, meta.InviteToken, "Guest")
	if err != nil {
		t.Fatal(err)
	}
	for name, tok := range map[string]string{
		"admin token": adminToken(t, s.srv),
		"guest seat":  playerToken(t, s.auth, meta.ID, playerID, "Guest"),
	} {
		for _, method := range []string{http.MethodGet, http.MethodPut} {
			if status, _ := notificationsCall(t, s, method, tok, map[string]bool{"turn": true}); status != http.StatusForbidden {
				t.Errorf("%s as %s: %d, want 403", method, name, status)
			}
		}
	}
}

// recordEvents is a TableEvents that remembers what it was told.
type recordEvents struct {
	mu      sync.Mutex
	filled  []uuid.UUID
	by      []uuid.UUID
	started []uuid.UUID
}

func (r *recordEvents) TableFilled(game, by uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.filled = append(r.filled, game)
	r.by = append(r.by, by)
}

func (r *recordEvents) TableStarted(game uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.started = append(r.started, game)
}

func TestTheLobbyReportsAFillAndAStart(t *testing.T) {
	l := NewLobby(ws.NewRoomManager(quietLogger(), ""))
	ev := &recordEvents{}
	l.SetTableEvents(ev)
	meta, err := l.Create("Pod")
	if err != nil {
		t.Fatal(err)
	}
	var last uuid.UUID
	for i, name := range []string{"A", "B", "C", "D"} {
		_, pid, err := l.Join(meta.ID, meta.InviteToken, name)
		if err != nil {
			t.Fatal(err)
		}
		last = pid
		if i < 3 && len(ev.filled) != 0 {
			t.Fatalf("a table with %d seats was reported full", i+1)
		}
	}
	if len(ev.filled) != 1 || ev.filled[0] != meta.ID || ev.by[0] != last {
		t.Fatalf("filled = %v by %v, want the table once, by the fourth seat %s", ev.filled, ev.by, last)
	}
	for _, s := range mustGet(t, l, meta.ID).Players {
		if _, err := l.SetDeck(meta.ID, s.PlayerID, "Deck", testDeck(s.Name)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	if len(ev.started) != 1 || ev.started[0] != meta.ID {
		t.Errorf("started = %v, want the table once", ev.started)
	}
}

func TestNotifyTablesShowsPeopleBotsAndTheHost(t *testing.T) {
	l := NewLobby(ws.NewRoomManager(quietLogger(), ""))
	meta, err := l.Create("Pod")
	if err != nil {
		t.Fatal(err)
	}
	user := uuid.New()
	_, alice, err := l.JoinAs(meta.ID, meta.InviteToken, "Alice", DiscordIdentity{}, user)
	if err != nil {
		t.Fatal(err)
	}
	_, guest, err := l.Join(meta.ID, meta.InviteToken, "Guest")
	if err != nil {
		t.Fatal(err)
	}
	nt := l.NotifyTables()
	if live := nt.Live(); len(live) != 0 {
		t.Errorf("Live lists a table still in the lobby: %+v", live)
	}
	tbl, ok := nt.Table(meta.ID)
	if !ok || tbl.State != notify.StateLobby || tbl.Full || tbl.Name != "Pod" {
		t.Fatalf("Table = %+v %v", tbl, ok)
	}
	bySeat := map[uuid.UUID]notify.Seat{}
	for _, s := range tbl.Seats {
		bySeat[s.PlayerID] = s
	}
	if s := bySeat[alice]; s.UserID != user || !s.Host {
		t.Errorf("Alice's seat = %+v, want her user and the host", s)
	}
	if s := bySeat[guest]; s.UserID != uuid.Nil || s.Host {
		t.Errorf("the guest's seat = %+v, want no user", s)
	}

	for _, s := range mustGet(t, l, meta.ID).Players {
		if _, err := l.SetDeck(meta.ID, s.PlayerID, "Deck", testDeck(s.Name)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	if live := nt.Live(); len(live) != 1 || live[0].ID != meta.ID {
		t.Fatalf("Live after the start = %+v", live)
	}
	// The opening roll waits on every seat that has not rolled.
	if waiting, acted := nt.Waiting(meta.ID, alice); !waiting || !acted.IsZero() {
		t.Errorf("Waiting(Alice) at the opening roll = %v, %v; want waiting, never acted", waiting, acted)
	}
	if waiting, _ := nt.Waiting(uuid.New(), alice); waiting {
		t.Error("an unknown table is waiting on someone")
	}
}

func TestNotifyTablesNeverShowsAPracticeTable(t *testing.T) {
	l, _, _ := newPracticeLobby(t)
	ev := &recordEvents{}
	l.SetTableEvents(ev)
	human, bot := practiceSeats("user:a")
	human.UserID = uuid.New()
	meta, _, err := l.CreatePractice(human, bot)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := l.NotifyTables().Table(meta.ID); ok {
		t.Error("the practice table was shown to the notifier")
	}
	for _, tbl := range l.NotifyTables().Live() {
		if tbl.ID == meta.ID {
			t.Error("the practice table is Live")
		}
	}
	if len(ev.started) != 0 || len(ev.filled) != 0 {
		t.Errorf("a practice table was reported: %+v", ev)
	}
}

func testDeck(name string) []game.Card {
	deck := []game.Card{game.NewCommander("Commander of "+name, uuid.Nil)}
	for range 20 {
		deck = append(deck, game.NewCard("Filler", uuid.Nil))
	}
	return deck
}
