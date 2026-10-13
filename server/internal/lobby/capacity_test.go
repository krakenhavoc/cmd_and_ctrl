package lobby

// capacity_test.go covers ADR 0148 (#2977): the weighted load cap, what
// it gates and what it never does, the friendly refusal on the wire,
// the per-account and per-IP creation rates, the connection check at
// the front door, and the numbers the admin Live now view and the
// metrics read.

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/capacity"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics/metricstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/ratelimit"
)

// weights used throughout: a table 2, a human seat 1, a bot seat 2, so
// a four-bot table is 10 and a four-human one is 6 (the shipped
// defaults).
func capLimits(maxLoad int) capacity.Limits {
	l := capacity.Defaults()
	l.MaxLoad = maxLoad
	return l
}

func refused(t *testing.T, reason string) float64 {
	t.Helper()
	return metricstest.Value(t, "cmdctrl_capacity_refusals_total", metricstest.L{"reason": reason})
}

// fillBots seats n bots at the table.
func fillBots(t *testing.T, l *Lobby, id uuid.UUID, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		if _, _, err := l.AddBot(id, "Bot", "random", "", "Mono Red", botDeck(20)); err != nil {
			t.Fatalf("AddBot %d: %v", i, err)
		}
	}
}

func TestTableCostWeighsABotSeatMoreThanAHumanSeat(t *testing.T) {
	lim := capacity.Defaults()
	cases := []struct {
		humans, bots, want int
	}{
		{0, 0, 2}, // an empty table still costs: the room, its writes, its memory
		{4, 0, 6}, // the owner's 30-game / 120-player shape: 30 x 6 = 180
		{0, 4, 10},
		{1, 3, 9},
	}
	for _, c := range cases {
		if got := lim.TableCost(c.humans, c.bots); got != c.want {
			t.Errorf("TableCost(%d humans, %d bots) = %d, want %d", c.humans, c.bots, got, c.want)
		}
	}
	// The owner's decision (#2977): 30 four-seat human tables, 120
	// players, fit the default cap exactly; the same 30 tables as bot
	// tables do not, and the cap holds 18 of them, well under the ~31
	// the load test collapsed at.
	if !lim.Fits(29*lim.TableCost(4, 0), lim.TableCost(4, 0)) {
		t.Error("the 30th four-human table must fit the default cap")
	}
	if lim.Fits(30*lim.TableCost(4, 0), lim.TableCost(4, 0)) {
		t.Error("a 31st four-human table must not fit the default cap")
	}
	if lim.Fits(17*lim.TableCost(0, 4), lim.TableCost(0, 4)) == false || lim.Fits(18*lim.TableCost(0, 4), lim.TableCost(0, 4)) {
		t.Error("the default cap must hold exactly 18 four-bot tables")
	}
}

func TestLoadCountsOnlyUnendedUnarchivedTables(t *testing.T) {
	l := newTestLobby(t)
	l.SetCapacity(capLimits(1000), nil)

	live, _ := l.Create("live")
	_, _, _ = l.Join(live.ID, live.InviteToken, "Alice")
	fillBots(t, l, live.ID, 2)
	if got, want := l.CapacityUsage().Load, 2+1+2*2; got != want {
		t.Fatalf("one human and two bots: load %d, want %d", got, want)
	}

	retired, _ := l.Create("retired")
	fillBots(t, l, retired.ID, 3)
	before := l.CapacityUsage().Load
	if _, err := l.SetArchived(retired.ID, true); err != nil {
		t.Fatal(err)
	}
	if got, want := l.CapacityUsage().Load, before-(2+3*2); got != want {
		t.Errorf("an archived table must cost nothing: load %d, want %d", got, want)
	}
}

func TestCreateRefusedAtTheCapAndExistingTablesKeepWorking(t *testing.T) {
	l := newTestLobby(t)
	// Room for exactly one empty table plus a little.
	l.SetCapacity(capLimits(2+2), nil)

	first, err := l.Create("first")
	if err != nil {
		t.Fatalf("Create under the cap: %v", err)
	}
	// Seats at the existing table push the load past the cap: that is
	// allowed, because the cap gates new load and never a seat at a
	// table that exists.
	_, alice, err := l.Join(first.ID, first.InviteToken, "Alice")
	if err != nil {
		t.Fatalf("a human joining an existing table was refused: %v", err)
	}
	if _, _, err := l.Join(first.ID, first.InviteToken, "Bob"); err != nil {
		t.Fatalf("second human: %v", err)
	}
	if got := l.CapacityUsage().Load; got < 4 {
		t.Fatalf("load %d should have reached the cap of 4", got)
	}

	before := refused(t, metrics.RefuseTableLoad)
	if _, err := l.Create("second"); !errors.Is(err, capacity.ErrServerFull) {
		t.Fatalf("Create over the cap: got %v, want ErrServerFull", err)
	}
	if refused(t, metrics.RefuseTableLoad) != before+1 {
		t.Error("a refused create must be counted")
	}
	// Nothing was created.
	if metas := l.List(); len(metas) != 1 {
		t.Errorf("a refused create left %d tables, want 1", len(metas))
	}

	// The existing table still works end to end: its deck upload and
	// start are not the cap's business.
	if _, err := l.SetDeck(first.ID, alice, "deck", botDeck(20)); err != nil {
		t.Errorf("SetDeck at an existing table: %v", err)
	}
	if _, err := l.Get(first.ID); err != nil {
		t.Errorf("Get: %v", err)
	}
}

func TestBotSeatRefusedAtTheCapButHumansAreNot(t *testing.T) {
	l := newTestLobby(t)
	// A table (2) + one bot (2) = 4 fits; a second bot (6) does not.
	l.SetCapacity(capLimits(4), nil)
	meta, err := l.Create("t")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.AddBot(meta.ID, "Bot", "random", "", "x", botDeck(20)); err != nil {
		t.Fatalf("the first bot fits: %v", err)
	}
	before := refused(t, metrics.RefuseBotSeat)
	_, _, err = l.AddBot(meta.ID, "Bot 2", "random", "", "x", botDeck(20))
	if !errors.Is(err, capacity.ErrServerFull) {
		t.Fatalf("the second bot: got %v, want ErrServerFull", err)
	}
	if refused(t, metrics.RefuseBotSeat) != before+1 {
		t.Error("a refused bot seat must be counted")
	}
	// The refused bot left no seat behind.
	got, _ := l.Get(meta.ID)
	if len(got.Players) != 1 {
		t.Errorf("seats after a refused bot: %d, want 1", len(got.Players))
	}
	// A human still sits down.
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Errorf("a human joining must never be refused by the load cap: %v", err)
	}
}

func TestPracticeTableRefusedAtTheCap(t *testing.T) {
	l, _, _ := newPracticeLobby(t)
	// A practice table is one human and one bot: 2 + 1 + 2 = 5.
	l.SetCapacity(capLimits(5), nil)
	h, b := practiceSeats("user:a")
	if _, _, err := l.CreatePractice(h, b); err != nil {
		t.Fatalf("the first practice table fits: %v", err)
	}
	h, b = practiceSeats("user:b")
	if _, _, err := l.CreatePractice(h, b); !errors.Is(err, capacity.ErrServerFull) {
		t.Fatalf("the second: got %v, want ErrServerFull", err)
	}
	// Replacing your own table frees what it costs first, so it fits.
	h, b = practiceSeats("user:a")
	if _, _, err := l.CreatePractice(h, b); err != nil {
		t.Errorf("an owner replacing their own table was refused: %v", err)
	}
}

// A burst of concurrent creates must not all pass the check: the
// reservation holds each one's cost until it is registered.
func TestConcurrentCreatesCannotOvershootTheCap(t *testing.T) {
	l := newTestLobby(t)
	const cap = 2 * 5 // five empty tables
	l.SetCapacity(capLimits(cap), nil)
	var ok, full atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := l.Create("race")
			switch {
			case err == nil:
				ok.Add(1)
			case errors.Is(err, capacity.ErrServerFull):
				full.Add(1)
			default:
				t.Errorf("Create: %v", err)
			}
		}()
	}
	wg.Wait()
	if ok.Load() != 5 || full.Load() != 35 {
		t.Errorf("created %d, refused %d; want 5 and 35", ok.Load(), full.Load())
	}
	if got := l.CapacityUsage().Load; got != cap {
		t.Errorf("load %d after the burst, want %d (the reservations must be released)", got, cap)
	}
}

func TestNoCapMeansNoLimit(t *testing.T) {
	l := newTestLobby(t) // SetCapacity never called: every cap off
	for i := 0; i < 50; i++ {
		meta, err := l.Create("t")
		if err != nil {
			t.Fatalf("Create %d: %v", i, err)
		}
		fillBots(t, l, meta.ID, 1)
	}
	if u := l.CapacityUsage(); u.LoadLimit != 0 || u.ConnLimit != 0 {
		t.Errorf("usage %+v: limits should read 0 (off)", u)
	}
}

func TestConnectionsFullGatesTheFrontDoorOnly(t *testing.T) {
	l := newTestLobby(t)
	conns := 0
	lim := capLimits(0)
	lim.MaxConns = 3
	l.SetCapacity(lim, func() int { return conns })

	conns = 2
	if l.ConnectionsFull() {
		t.Error("2 of 3: not full")
	}
	conns = 3
	before := refused(t, metrics.RefuseJoinConnections)
	if !l.ConnectionsFull() {
		t.Error("3 of 3: full")
	}
	if refused(t, metrics.RefuseJoinConnections) != before+1 {
		t.Error("a full check that refuses must be counted")
	}
	if u := l.CapacityUsage(); u.Conns != 3 || u.ConnLimit != 3 {
		t.Errorf("usage %+v", u)
	}
}

// --- over HTTP --------------------------------------------------------

type capStack struct {
	srv   *httptest.Server
	lobby *Lobby
	auth  auth.Authenticator
}

func newCapStack(t *testing.T, lim capacity.Limits, conns func() int) capStack {
	t.Helper()
	a := auth.NewMemoryAuthenticator()
	mgrLobby := newTestLobby(t)
	mgrLobby.SetCapacity(lim, conns)
	cfg := Config{Lobby: mgrLobby, Auth: a, AdminToken: "shared-admin-token", Capacity: lim}
	srv := httptest.NewServer(Handler(cfg))
	t.Cleanup(srv.Close)
	return capStack{srv, mgrLobby, a}
}

// decodeFull asserts resp is the friendly capacity refusal and nothing
// raw: 503, the stable code, the player-facing sentence, a Retry-After.
func decodeFull(t *testing.T, resp *http.Response) {
	t.Helper()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status %d, want 503", resp.StatusCode)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != "server_full" {
		t.Errorf("code %q, want the stable server_full", body["code"])
	}
	if capacity.CodeServerFull != "server_full" {
		t.Error("the stable code changed; the client matches the literal")
	}
	if body["error"] != capacity.Message {
		t.Errorf("error %q, want the friendly message", body["error"])
	}
	if resp.Header.Get("Retry-After") == "" {
		t.Error("a capacity refusal must say when to retry")
	}
}

func TestFullServerAnswersACreateWithAStableCode(t *testing.T) {
	s := newCapStack(t, capLimits(2), nil)
	tok := adminToken(t, s.srv)

	resp := postJSON(t, s.srv, "/games", tok, createGameRequest{Name: "one"})
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("the first create: %d", resp.StatusCode)
	}
	decodeFull(t, postJSON(t, s.srv, "/games", tok, createGameRequest{Name: "two"}))
}

func TestFullServerAnswersABotSeatWithAStableCode(t *testing.T) {
	srv, l, _, _ := newBotHTTPStack(t, buildMinimalDeckIndex(t), fakeDeckSource{})
	meta, _ := l.Create("FNM")
	joined := joinAs(t, srv, meta, "Alice")
	// The table with Alice is 2 + 1 = 3. One more bot makes 5.
	l.SetCapacity(capLimits(4), nil)

	decodeFull(t, postJSON(t, srv, "/games/"+meta.ID.String()+"/seats/bot", joined.Token,
		addBotRequest{Tier: "random", Deck: "test-mono-white"}))
	if got, _ := l.Get(meta.ID); len(got.Players) != 1 {
		t.Errorf("a refused bot left %d seats, want Alice alone", len(got.Players))
	}

	// With room it is seated, so the refusal was the cap and nothing else.
	l.SetCapacity(capLimits(5), nil)
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/seats/bot", joined.Token,
		addBotRequest{Tier: "random", Deck: "test-mono-white"})
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("bot with room: %d, want 201", resp.StatusCode)
	}
}

func TestFullConnectionsRefuseAJoinAndASpectateButNotTheAdmin(t *testing.T) {
	conns := 5
	lim := capLimits(0)
	lim.MaxConns = 5
	s := newCapStack(t, lim, func() int { return conns })
	meta, err := s.lobby.Create("t")
	if err != nil {
		t.Fatal(err)
	}

	decodeFull(t, postJSON(t, s.srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"}))
	decodeFull(t, postJSON(t, s.srv, "/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"}))
	decodeFull(t, postJSON(t, s.srv, "/games/"+meta.ID.String()+"/spectate", "",
		spectateRequest{InviteToken: meta.SpectatorInvite}))

	// The refusal seated nobody.
	if got, _ := s.lobby.Get(meta.ID); len(got.Players) != 0 {
		t.Errorf("a refused join left %d seats", len(got.Players))
	}
	// Once there is room again, the same join works.
	conns = 4
	resp := postJSON(t, s.srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("join with room: %d, want 200", resp.StatusCode)
	}
}

func signedIn(t *testing.T, s capStack) string {
	t.Helper()
	return identifiedToken(t, s.auth, uuid.New())
}

func TestPerAccountCreationRateIsConfigurable(t *testing.T) {
	lim := capLimits(0)
	lim.AccountBurst = 2
	lim.AccountEvery = time.Hour
	lim.IPBurst = 100 // out of the way
	s := newCapStack(t, lim, nil)
	alice := signedIn(t, s)
	bob := signedIn(t, s)

	before := refused(t, metrics.RefuseCreateAccount)
	for i := 0; i < 2; i++ {
		resp := postJSON(t, s.srv, "/games", alice, createGameRequest{Name: "t"})
		resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("create %d: %d, want 201", i, resp.StatusCode)
		}
	}
	// The creator also has an open-table cap of three; two are open.
	resp := postJSON(t, s.srv, "/games", alice, createGameRequest{Name: "t"})
	resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("third create from one account: %d, want 429", resp.StatusCode)
	}
	if resp.Header.Get("Retry-After") == "" {
		t.Error("a rate refusal must say when to retry")
	}
	if refused(t, metrics.RefuseCreateAccount) != before+1 {
		t.Error("an account-rate refusal must be counted")
	}
	// Another account has its own bucket.
	resp = postJSON(t, s.srv, "/games", bob, createGameRequest{Name: "t"})
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("another account: %d, want 201", resp.StatusCode)
	}
}

func TestPerIPCreationRateBoundsManyAccountsFromOneAddress(t *testing.T) {
	lim := capLimits(0)
	lim.AccountBurst = 100
	lim.IPBurst = 2
	lim.IPEvery = time.Hour
	s := newCapStack(t, lim, nil)

	before := refused(t, metrics.RefuseCreateIP)
	// Three different accounts from one address (httptest: one
	// loopback IP): the third is held by the address, not the account.
	for i := 0; i < 2; i++ {
		resp := postJSON(t, s.srv, "/games", signedIn(t, s), createGameRequest{Name: "t"})
		resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("create %d: %d, want 201", i, resp.StatusCode)
		}
	}
	resp := postJSON(t, s.srv, "/games", signedIn(t, s), createGameRequest{Name: "t"})
	resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("third create from one address: %d, want 429", resp.StatusCode)
	}
	if refused(t, metrics.RefuseCreateIP) != before+1 {
		t.Error("an IP-rate refusal must be counted")
	}
	// The shared admin token (the Discord bot) is not held to it.
	resp = postJSON(t, s.srv, "/games", adminToken(t, s.srv), createGameRequest{Name: "bot"})
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("the shared token: %d, want 201", resp.StatusCode)
	}
}

func TestIPGateBoundsAGuestsPracticeTables(t *testing.T) {
	// createIPGate is the practice route's per-IP bucket (a guest
	// cannot create a table, but can open practice tables).
	lim := capacity.Defaults()
	lim.IPBurst = 1
	lim.IPEvery = time.Hour
	rl := ratelimit.New(1/lim.IPEvery.Seconds(), float64(lim.IPBurst))
	var served int
	h := createIPGate(rl, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { served++ }))

	do := func() *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/games/practice", nil))
		return rec
	}
	if rec := do(); rec.Code != http.StatusOK || served != 1 {
		t.Fatalf("first: %d served %d", rec.Code, served)
	}
	rec := do()
	if rec.Code != http.StatusTooManyRequests || served != 1 {
		t.Fatalf("second: %d served %d, want 429 and not served", rec.Code, served)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("missing Retry-After")
	}
}

func TestAdminLiveShowsTheCaps(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	s := newLiveNowStack(t, nil)
	lim := capLimits(180)
	lim.MaxConns = 400
	s.lobby.SetCapacity(lim, func() int { return 7 })
	meta, err := s.lobby.Create("t")
	if err != nil {
		t.Fatal(err)
	}
	fillBots(t, s.lobby, meta.ID, 2)

	tok := adminToken(t, s.srv)
	var got liveNowResponse
	if err := json.Unmarshal(s.live(t, tok), &got); err != nil {
		t.Fatal(err)
	}
	want := liveNowCapacity{Load: 2 + 2*2, LoadLimit: 180, Connections: 7, ConnectionLimit: 400}
	if got.Capacity != want {
		t.Errorf("capacity %+v, want %+v", got.Capacity, want)
	}
}

// The capacity collector reports the same numbers the admin view does.
func TestCapacityCollectorReportsLoadAndLimits(t *testing.T) {
	l := newTestLobby(t)
	lim := capLimits(50)
	lim.MaxConns = 9
	l.SetCapacity(lim, func() int { return 4 })
	meta, _ := l.Create("t")
	fillBots(t, l, meta.ID, 1)

	r := l.CapacityReading()
	if r.Load != 4 || r.LoadLimit != 50 || r.Conns != 4 || r.ConnLimit != 9 {
		t.Errorf("reading %+v", r)
	}
}
