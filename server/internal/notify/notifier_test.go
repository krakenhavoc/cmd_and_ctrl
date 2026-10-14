package notify

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

// The Notifier against fakes: a fake clock, fake tables, a fake hub and
// a fake DM sender. Nothing here reaches Discord except the one 50007
// test, which drives the real discord.Bot against an httptest server.

var t0 = time.Date(2026, 10, 13, 20, 0, 0, 0, time.UTC)

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time          { return c.now }
func (c *fakeClock) advance(d time.Duration) { c.now = c.now.Add(d) }

type fakeTables struct {
	tables  map[uuid.UUID]*Table
	waiting map[seatKey]bool
	acted   map[seatKey]time.Time
}

func (f *fakeTables) Live() []Table {
	var out []Table
	for _, t := range f.tables {
		if t.State == StateActive {
			out = append(out, *t)
		}
	}
	return out
}

func (f *fakeTables) Table(id uuid.UUID) (Table, bool) {
	t, ok := f.tables[id]
	if !ok {
		return Table{}, false
	}
	return *t, true
}

func (f *fakeTables) Waiting(game, seat uuid.UUID) (bool, time.Time) {
	k := seatKey{game, seat}
	return f.waiting[k], f.acted[k]
}

type presence struct {
	away  bool
	since time.Time
}

type fakeHub struct{ seats map[seatKey]presence }

func (h *fakeHub) SeatAway(game, seat uuid.UUID) (bool, time.Time) {
	p, ok := h.seats[seatKey{game, seat}]
	if !ok {
		return true, time.Time{}
	}
	return p.away, p.since
}

type memStore struct {
	mu    sync.Mutex
	prefs map[uuid.UUID]Prefs
}

func (s *memStore) Get(_ context.Context, u uuid.UUID) (Prefs, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.prefs[u], nil
}

func (s *memStore) Put(context.Context, uuid.UUID, Update) (Prefs, error) {
	return Prefs{}, ErrNoStore
}

func (s *memStore) Active(context.Context) (map[uuid.UUID]Prefs, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[uuid.UUID]Prefs)
	for u, p := range s.prefs {
		if p.DMFailedAt.IsZero() && (p.Turn || p.TableReady || p.TableFilled) {
			out[u] = p
		}
	}
	return out, nil
}

func (s *memStore) MarkFailed(_ context.Context, u uuid.UUID, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.prefs[u]
	p.DMFailedAt = at
	s.prefs[u] = p
	return nil
}

type subjects map[uuid.UUID]string

func (s subjects) DiscordSubject(_ context.Context, u uuid.UUID) (string, error) {
	if id, ok := s[u]; ok {
		return id, nil
	}
	return "", fmt.Errorf("no identity")
}

type sent struct{ to, content string }

type fakeSender struct {
	sent []sent
	// errs are returned in order, one per call, before calls succeed.
	errs []error
}

func (f *fakeSender) SendDM(_ context.Context, to, content string) error {
	if len(f.errs) > 0 {
		err := f.errs[0]
		f.errs = f.errs[1:]
		if err != nil {
			return err
		}
	}
	f.sent = append(f.sent, sent{to, content})
	return nil
}

// world is one table with four seats: Alice (a signed-in person who
// hosts), a bot, an agent seat and a guest.
type world struct {
	clock  *fakeClock
	tables *fakeTables
	hub    *fakeHub
	store  *memStore
	sender *fakeSender
	n      *Notifier
	slept  []time.Duration

	game                    uuid.UUID
	alice, bot, agent, gues Seat
}

func newWorld(t *testing.T) *world {
	t.Helper()
	w := &world{
		clock:  &fakeClock{now: t0},
		hub:    &fakeHub{seats: map[seatKey]presence{}},
		store:  &memStore{prefs: map[uuid.UUID]Prefs{}},
		sender: &fakeSender{},
		game:   uuid.New(),
	}
	w.alice = Seat{PlayerID: uuid.New(), UserID: uuid.New(), Host: true}
	w.bot = Seat{PlayerID: uuid.New(), Bot: true}
	w.agent = Seat{PlayerID: uuid.New(), Agent: true}
	w.gues = Seat{PlayerID: uuid.New()}
	w.tables = &fakeTables{
		tables: map[uuid.UUID]*Table{w.game: {
			ID: w.game, Name: "Friday *pod*", State: StateActive, Turn: 7, Full: true,
			Seats: []Seat{w.alice, w.bot, w.agent, w.gues},
		}},
		waiting: map[seatKey]bool{},
		acted:   map[seatKey]time.Time{},
	}
	w.n = New(Config{
		Tables:     w.tables,
		Presence:   w.hub,
		Store:      w.store,
		Recipients: subjects{w.alice.UserID: "snowflake-alice"},
		Sender:     w.sender,
		BaseURL:    "https://cmd.example/",
		Now:        w.clock.Now,
		Sleep: func(_ context.Context, d time.Duration) error {
			w.slept = append(w.slept, d)
			return nil
		},
	})
	return w
}

func (w *world) key(s Seat) seatKey { return seatKey{w.game, s.PlayerID} }

// tickFor advances the clock in TickEvery steps over d, ticking each
// step, as Run would.
func (w *world) tickFor(d time.Duration) {
	for end := w.clock.now.Add(d); w.clock.now.Before(end); {
		w.clock.advance(TickEvery)
		w.n.Tick(context.Background())
	}
}

func (w *world) tick() { w.n.Tick(context.Background()) }

func TestNothingIsSentToSomeoneWhoNeverOptedIn(t *testing.T) {
	w := newWorld(t)
	w.tables.waiting[w.key(w.alice)] = true
	w.tick()
	w.tickFor(30 * time.Minute)
	w.n.TableStarted(w.game)
	w.tickFor(time.Minute)
	if len(w.sender.sent) != 0 {
		t.Fatalf("sent %v to a player with every switch off", w.sender.sent)
	}
}

func TestYourTurnAfterTwoMinutesAway(t *testing.T) {
	w := newWorld(t)
	w.store.prefs[w.alice.UserID] = Prefs{Turn: true}
	w.tables.waiting[w.key(w.alice)] = true
	w.tick()
	w.tickFor(TurnWait - TickEvery)
	if len(w.sender.sent) != 0 {
		t.Fatalf("sent before two minutes: %v", w.sender.sent)
	}
	w.tickFor(TickEvery)
	if len(w.sender.sent) != 1 {
		t.Fatalf("after two minutes away, sent %d DMs, want 1", len(w.sender.sent))
	}
	got := w.sender.sent[0]
	want := "It's your turn at **Friday pod** (turn 7). https://cmd.example/#/games/" + w.game.String()
	if got.to != "snowflake-alice" || got.content != want {
		t.Errorf("sent %+v, want %q to snowflake-alice", got, want)
	}
	// Still waiting, still away: one DM for the stretch, not one a tick.
	w.tickFor(30 * time.Minute)
	if len(w.sender.sent) != 1 {
		t.Errorf("one unbroken wait sent %d DMs, want 1", len(w.sender.sent))
	}
}

func TestActingResetsTheClock(t *testing.T) {
	w := newWorld(t)
	w.store.prefs[w.alice.UserID] = Prefs{Turn: true}
	w.tables.waiting[w.key(w.alice)] = true
	w.tick()
	w.tickFor(90 * time.Second)
	// Alice acts (cast a spell and kept priority) from a tab that then
	// went hidden again.
	w.tables.acted[w.key(w.alice)] = w.clock.now
	w.tickFor(TurnWait - TickEvery)
	if len(w.sender.sent) != 0 {
		t.Fatalf("sent within two minutes of the player acting: %v", w.sender.sent)
	}
	w.tickFor(TickEvery)
	if len(w.sender.sent) != 1 {
		t.Errorf("two minutes after acting, sent %d, want 1", len(w.sender.sent))
	}
}

func TestTheClockStopsWhenTheEngineStopsWaiting(t *testing.T) {
	w := newWorld(t)
	w.store.prefs[w.alice.UserID] = Prefs{Turn: true}
	w.tables.waiting[w.key(w.alice)] = true
	w.tick()
	w.tickFor(90 * time.Second)
	w.tables.waiting[w.key(w.alice)] = false
	w.tickFor(TickEvery)
	w.tables.waiting[w.key(w.alice)] = true
	w.tickFor(90 * time.Second)
	if len(w.sender.sent) != 0 {
		t.Fatalf("two separate 90-second waits added up to a DM: %v", w.sender.sent)
	}
}

// ADR 0162 §2: away means no socket, or every socket hidden. The hub
// answers that (ws.Hub.SeatAway, tested there); here, a present seat is
// never messaged, and coming back starts the two minutes again.
func TestPresentNeverAndComingBackResets(t *testing.T) {
	w := newWorld(t)
	w.store.prefs[w.alice.UserID] = Prefs{Turn: true}
	w.tables.waiting[w.key(w.alice)] = true
	w.hub.seats[w.key(w.alice)] = presence{away: false}
	w.tick()
	w.tickFor(10 * time.Minute)
	if len(w.sender.sent) != 0 {
		t.Fatalf("sent to a player who is at the table: %v", w.sender.sent)
	}
	// Alice hides the tab: the hub stamps when.
	w.hub.seats[w.key(w.alice)] = presence{away: true, since: w.clock.now}
	w.tickFor(time.Minute)
	// Back for a moment, and away again.
	w.hub.seats[w.key(w.alice)] = presence{away: true, since: w.clock.now}
	w.tickFor(TurnWait - TickEvery)
	if len(w.sender.sent) != 0 {
		t.Fatalf("sent within two minutes of the player last leaving: %v", w.sender.sent)
	}
	w.tickFor(TickEvery)
	if len(w.sender.sent) != 1 {
		t.Errorf("two minutes after leaving, sent %d, want 1", len(w.sender.sent))
	}
}

func TestNeverToBotsAgentsOrGuests(t *testing.T) {
	w := newWorld(t)
	// Even with a user and every switch on, a bot or agent seat is
	// never messaged; a guest has no user to message.
	botUser, agentUser := uuid.New(), uuid.New()
	tbl := w.tables.tables[w.game]
	tbl.Seats[1].UserID = botUser
	tbl.Seats[2].UserID = agentUser
	all := Prefs{Turn: true, TableReady: true, TableFilled: true}
	w.store.prefs[botUser] = all
	w.store.prefs[agentUser] = all
	w.n.cfg.Recipients = subjects{botUser: "snowflake-bot", agentUser: "snowflake-agent"}
	for _, s := range tbl.Seats {
		w.tables.waiting[w.key(s)] = true
	}
	w.n.TableStarted(w.game)
	w.tick()
	w.tickFor(30 * time.Minute)
	if len(w.sender.sent) != 0 {
		t.Fatalf("sent %v to a bot, an agent or a guest", w.sender.sent)
	}
}

// One DM per player per table per ten minutes, across every kind.
func TestTenMinuteWindowAcrossKinds(t *testing.T) {
	w := newWorld(t)
	w.store.prefs[w.alice.UserID] = Prefs{Turn: true, TableReady: true}
	w.n.TableStarted(w.game)
	w.tickFor(EventGrace)
	if len(w.sender.sent) != 1 || !strings.Contains(w.sender.sent[0].content, "has started") {
		t.Fatalf("the start sent %v, want one \"has started\"", w.sender.sent)
	}
	// The game then waits on Alice, who is still away.
	w.tables.waiting[w.key(w.alice)] = true
	w.tickFor(5 * time.Minute)
	if len(w.sender.sent) != 1 {
		t.Fatalf("a turn DM inside the window: %v", w.sender.sent)
	}
	// The stretch the window swallowed is not sent late, even once the
	// window is up.
	w.tickFor(6 * time.Minute)
	if len(w.sender.sent) != 1 {
		t.Fatalf("a held-back turn DM went out late: %v", w.sender.sent)
	}
	// Acting starts a new stretch, which may send now the window is up.
	w.tables.acted[w.key(w.alice)] = w.clock.now
	w.tickFor(TurnWait)
	if len(w.sender.sent) != 2 || !strings.Contains(w.sender.sent[1].content, "your turn") {
		t.Errorf("after the window, sent %v, want the turn DM second", w.sender.sent)
	}
}

func TestTheWindowIsPerTable(t *testing.T) {
	w := newWorld(t)
	w.store.prefs[w.alice.UserID] = Prefs{TableReady: true}
	other := uuid.New()
	w.tables.tables[other] = &Table{ID: other, Name: "Other", State: StateActive, Seats: []Seat{w.alice}}
	w.n.TableStarted(w.game)
	w.n.TableStarted(other)
	w.tickFor(EventGrace)
	if len(w.sender.sent) != 2 {
		t.Errorf("two tables starting sent %d DMs, want one each", len(w.sender.sent))
	}
}

func TestFilledTellsTheHostAndNotTheJoiner(t *testing.T) {
	w := newWorld(t)
	tbl := w.tables.tables[w.game]
	tbl.State = StateLobby
	bea := Seat{PlayerID: uuid.New(), UserID: uuid.New()}
	cal := Seat{PlayerID: uuid.New(), UserID: uuid.New()}
	tbl.Seats = []Seat{w.alice, bea, cal, w.bot}
	w.store.prefs[w.alice.UserID] = Prefs{TableFilled: true}
	w.store.prefs[bea.UserID] = Prefs{TableReady: true}
	w.store.prefs[cal.UserID] = Prefs{TableReady: true, TableFilled: true}
	w.n.cfg.Recipients = subjects{w.alice.UserID: "a", bea.UserID: "b", cal.UserID: "c"}
	// Cal's join filled the table.
	w.n.TableFilled(w.game, cal.PlayerID)
	w.tickFor(EventGrace - TickEvery)
	if len(w.sender.sent) != 0 {
		t.Fatalf("sent inside the grace: %v", w.sender.sent)
	}
	w.tickFor(TickEvery)
	to := map[string]string{}
	for _, s := range w.sender.sent {
		to[s.to] = s.content
	}
	if len(to) != 2 || to["a"] == "" || to["b"] == "" {
		t.Fatalf("sent to %v, want the host and Bea, not Cal who filled it", to)
	}
	if !strings.Contains(to["a"], "is full and ready to start") {
		t.Errorf("the host's DM = %q", to["a"])
	}
}

func TestFilledIsDroppedOnceTheTableHasStarted(t *testing.T) {
	w := newWorld(t)
	w.tables.tables[w.game].State = StateLobby
	w.store.prefs[w.alice.UserID] = Prefs{TableFilled: true}
	w.n.TableFilled(w.game, uuid.Nil)
	w.tables.tables[w.game].State = StateActive
	w.tickFor(EventGrace)
	if len(w.sender.sent) != 0 {
		t.Errorf("a fill was reported for a table already in play: %v", w.sender.sent)
	}
}

func TestEventsSkipAPlayerWhoIsPresent(t *testing.T) {
	w := newWorld(t)
	w.store.prefs[w.alice.UserID] = Prefs{TableReady: true}
	w.hub.seats[w.key(w.alice)] = presence{away: false}
	w.n.TableStarted(w.game)
	w.tickFor(EventGrace)
	if len(w.sender.sent) != 0 {
		t.Errorf("sent %v to a player at the table", w.sender.sent)
	}
}

// Discord's 50007 records the refusal and stops every later DM.
func TestRefusalStopsSending(t *testing.T) {
	w := newWorld(t)
	w.store.prefs[w.alice.UserID] = Prefs{Turn: true, TableReady: true}
	w.sender.errs = []error{fmt.Errorf("open: %w", discord.ErrDMRefused)}
	w.n.TableStarted(w.game)
	w.tickFor(EventGrace)
	if p := w.store.prefs[w.alice.UserID]; !p.DMFailedAt.Equal(w.clock.now) {
		t.Fatalf("after 50007 the refusal is %v, want it recorded at %v", p.DMFailedAt, w.clock.now)
	}
	w.tables.waiting[w.key(w.alice)] = true
	w.tickFor(time.Hour)
	if len(w.sender.sent) != 0 {
		t.Errorf("sent %v after Discord refused", w.sender.sent)
	}
}

// The same, through the real discord.Bot: a 403 with code 50007.
func TestRefusalFromDiscordItself(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		calls++
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusForbidden)
		_, _ = rw.Write([]byte(`{"message":"Cannot send messages to this user","code":50007}`))
	}))
	defer srv.Close()
	defer discord.SwapBotAPIBaseForTesting(srv.URL)()

	w := newWorld(t)
	w.n.cfg.Sender = BotSender{Bot: discord.Bot{Token: "tok"}, Client: srv.Client()}
	w.store.prefs[w.alice.UserID] = Prefs{TableReady: true}
	w.n.TableStarted(w.game)
	w.tickFor(EventGrace)
	if calls != 1 {
		t.Fatalf("Discord was called %d times, want 1", calls)
	}
	if w.store.prefs[w.alice.UserID].DMFailedAt.IsZero() {
		t.Error("a 403 with code 50007 did not record the refusal")
	}
}

func TestRateLimitWaitsOnceThenDrops(t *testing.T) {
	w := newWorld(t)
	w.store.prefs[w.alice.UserID] = Prefs{TableReady: true}
	w.sender.errs = []error{discord.ErrDMRateLimited}
	w.n.TableStarted(w.game)
	w.tickFor(EventGrace)
	if len(w.slept) != 1 || len(w.sender.sent) != 1 {
		t.Fatalf("one 429: slept %v and sent %d, want one wait and the DM", w.slept, len(w.sender.sent))
	}

	w2 := newWorld(t)
	w2.store.prefs[w2.alice.UserID] = Prefs{TableReady: true}
	w2.sender.errs = []error{discord.ErrDMRateLimited, discord.ErrDMRateLimited}
	w2.n.TableStarted(w2.game)
	w2.tickFor(EventGrace)
	w2.tickFor(time.Hour)
	if len(w2.slept) != 1 || len(w2.sender.sent) != 0 {
		t.Errorf("two 429s: slept %v and sent %d, want one wait, then dropped", w2.slept, len(w2.sender.sent))
	}
	if !w2.store.prefs[w2.alice.UserID].DMFailedAt.IsZero() {
		t.Error("a 429 was recorded as a refusal")
	}
}

func TestNeverRetriedLate(t *testing.T) {
	w := newWorld(t)
	w.store.prefs[w.alice.UserID] = Prefs{Turn: true}
	w.sender.errs = []error{fmt.Errorf("discord: DM failed: status 500")}
	w.tables.waiting[w.key(w.alice)] = true
	w.tick()
	w.tickFor(TurnWait)
	w.tickFor(30 * time.Minute)
	if len(w.sender.sent) != 0 {
		t.Errorf("a failed turn DM was sent again later: %v", w.sender.sent)
	}
}

func TestTurnMessageBeforeTurnOne(t *testing.T) {
	if got := turnMessage("Pod", 0, "L"); got != "The game at **Pod** is waiting on you. L" {
		t.Errorf("turnMessage before turn 1 = %q", got)
	}
	if got := flatten("  a\n*b*  @everyone  "); got != "a b everyone" {
		t.Errorf("flatten = %q", got)
	}
}
