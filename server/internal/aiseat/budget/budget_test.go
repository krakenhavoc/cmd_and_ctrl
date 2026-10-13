package budget

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// budget_test.go is #2980's proof: a fake model client that reports
// usage, the cap tripping, the fallback to Layer B, the kill switch,
// persistence across a "restart", and the day and month roll-over.

func quiet() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// clock is a settable wall clock.
type clock struct {
	mu sync.Mutex
	t  time.Time
}

func newClock(s string) *clock {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return &clock{t: t}
}
func (c *clock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}
func (c *clock) set(s string) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	c.mu.Lock()
	c.t = t
	c.mu.Unlock()
}

// usageClient is a fake model client that reports a fixed usage per
// call, so a test can say exactly how many calls fit under a cap.
func usageClient(in, out int) *model.FakeClient {
	return &model.FakeClient{
		Label: "usage",
		Reply: func(int, model.Request) (model.Response, error) {
			return model.Response{
				Text:  `{"index": 1, "why": "fake"}`,
				Usage: model.Usage{InputTokens: in, OutputTokens: out},
			}, nil
		},
	}
}

func call(t *testing.T, c model.Client) error {
	t.Helper()
	_, err := c.Complete(context.Background(), model.Request{Model: "m"})
	return err
}

func TestTokensCountsEveryBilledField(t *testing.T) {
	got := Tokens(model.Usage{InputTokens: 1, OutputTokens: 2, CacheReadTokens: 4, CacheWriteTokens: 8, CachedPromptTokens: 1000})
	if got != 15 {
		t.Fatalf("Tokens = %d, want 15 (CachedPromptTokens is a local server's free optimisation and does not count)", got)
	}
}

func TestDailyCapTripsAndLaterCallsNeverReachTheInnerClient(t *testing.T) {
	inner := usageClient(60, 40) // 100 tokens a call
	b := New(Limits{DailyTokens: 250}, nil, WithLogger(quiet()))
	c := b.Wrap(inner)

	for i := 0; i < 3; i++ { // 100, 200, 300: the third call crosses the line
		if err := call(t, c); err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
	}
	if st := b.Status(); st.State != StateDailyCap || st.DailyTokens != 300 {
		t.Fatalf("status = %+v, want daily_cap at 300", st)
	}
	err := call(t, c)
	if !errors.Is(err, model.ErrBudgetExceeded) {
		t.Fatalf("call over the cap: err = %v, want ErrBudgetExceeded", err)
	}
	if inner.Calls() != 3 {
		t.Fatalf("inner client saw %d calls, want 3: a refused call must not leave the process", inner.Calls())
	}
	if b.Blocked() == "" || !strings.Contains(b.Blocked(), "daily") {
		t.Fatalf("Blocked() = %q, want the daily-cap note", b.Blocked())
	}
}

func TestMonthlyCapTripsAcrossDays(t *testing.T) {
	clk := newClock("2026-10-10T12:00:00Z")
	b := New(Limits{DailyTokens: 1000, MonthlyTokens: 250}, nil, WithClock(clk.now), WithLogger(quiet()))
	c := b.Wrap(usageClient(100, 0))
	for i := 0; i < 2; i++ {
		_ = call(t, c)
	}
	clk.set("2026-10-11T12:00:00Z") // new day, same month: 200 so far
	if st := b.Status(); st.DailyTokens != 0 || st.MonthlyTokens != 200 || st.State != StateOK {
		t.Fatalf("after the day rolls: %+v", st)
	}
	if err := call(t, c); err != nil {
		t.Fatal(err)
	}
	st := b.Status()
	if st.State != StateMonthlyCap {
		t.Fatalf("state = %s, want monthly_cap", st.State)
	}
	if !strings.Contains(st.Note, "monthly") || !strings.Contains(st.Note, "2026-11-01 00:00 UTC") {
		t.Fatalf("note = %q, want the monthly note naming the 1st", st.Note)
	}
}

func TestDayAndMonthRollOver(t *testing.T) {
	clk := newClock("2026-10-31T23:59:00Z")
	b := New(Limits{DailyTokens: 100, MonthlyTokens: 100}, nil, WithClock(clk.now), WithLogger(quiet()))
	c := b.Wrap(usageClient(100, 0))
	_ = call(t, c)
	if err := call(t, c); !errors.Is(err, model.ErrBudgetExceeded) {
		t.Fatalf("before midnight: err = %v, want the cap", err)
	}

	clk.set("2026-11-01T00:00:01Z") // both boundaries at once
	if st := b.Status(); st.DailyTokens != 0 || st.MonthlyTokens != 0 || st.State != StateOK {
		t.Fatalf("after midnight on the 1st: %+v, want both counters reset", st)
	}
	if err := call(t, c); err != nil {
		t.Fatalf("after the reset: %v", err)
	}
}

func TestZeroMeansNoCap(t *testing.T) {
	b := New(Limits{}, nil, WithLogger(quiet()))
	c := b.Wrap(usageClient(1_000_000, 1_000_000))
	for i := 0; i < 5; i++ {
		if err := call(t, c); err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
	}
	if st := b.Status(); st.State != StateOK || st.DailyTokens != 10_000_000 {
		t.Fatalf("status = %+v: a zero cap still counts, it just never trips", st)
	}
}

func TestKillSwitchStopsCallsAndSurvivesTurningBackOn(t *testing.T) {
	inner := usageClient(1, 1)
	b := New(DefaultLimits(), nil, WithLogger(quiet()))
	c := b.Wrap(inner)
	if err := call(t, c); err != nil {
		t.Fatal(err)
	}
	if err := b.SetEnabled(false); err != nil {
		t.Fatal(err)
	}
	if err := call(t, c); !errors.Is(err, model.ErrBudgetExceeded) {
		t.Fatalf("switched off: err = %v", err)
	}
	if st := b.Status(); st.State != StateDisabled || st.Enabled {
		t.Fatalf("status = %+v, want disabled", st)
	}
	if err := b.SetEnabled(true); err != nil {
		t.Fatal(err)
	}
	if err := call(t, c); err != nil {
		t.Fatalf("switched back on: %v", err)
	}
	if inner.Calls() != 2 {
		t.Fatalf("inner calls = %d, want 2 (the switched-off call must not leave the process)", inner.Calls())
	}
}

// memStore is an in-memory Store so the persistence contract is
// testable without a database; the SQL one has its own test.
type memStore struct {
	mu       sync.Mutex
	day      string
	dayN     int64
	month    string
	monthN   int64
	disabled bool
	failSave bool
}

func (m *memStore) Load(_ context.Context, day, month string) (Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := Snapshot{Disabled: m.disabled}
	if m.day == day {
		s.DayTokens = m.dayN
	}
	if m.month == month {
		s.MonthTokens = m.monthN
	}
	return s, nil
}
func (m *memStore) SaveTokens(_ context.Context, day string, dn int64, month string, mn int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failSave {
		return errors.New("disk full")
	}
	m.day, m.dayN, m.month, m.monthN = day, dn, month, mn
	return nil
}
func (m *memStore) SaveDisabled(_ context.Context, d bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.disabled = d
	return nil
}

func TestCountersAndSwitchSurviveARestart(t *testing.T) {
	clk := newClock("2026-10-12T09:00:00Z")
	store := &memStore{}
	b1 := New(Limits{DailyTokens: 150}, store, WithClock(clk.now), WithLogger(quiet()))
	c1 := b1.Wrap(usageClient(100, 0))
	_ = call(t, c1)
	_ = call(t, c1) // 200: capped
	if err := b1.SetEnabled(false); err != nil {
		t.Fatal(err)
	}

	// "Restart": a new Budget over the same store, same day.
	b2 := New(Limits{DailyTokens: 150}, store, WithClock(clk.now), WithLogger(quiet()))
	if st := b2.Status(); st.DailyTokens != 200 || st.Enabled {
		t.Fatalf("after restart: %+v, want 200 tokens and the switch still off", st)
	}
	_ = b2.SetEnabled(true)
	if err := call(t, b2.Wrap(usageClient(1, 1))); !errors.Is(err, model.ErrBudgetExceeded) {
		t.Fatalf("a restart must not hand a capped server a fresh allowance: err = %v", err)
	}

	// Restart the next day: the stored day belongs to yesterday.
	clk.set("2026-10-13T00:00:05Z")
	b3 := New(Limits{DailyTokens: 150}, store, WithClock(clk.now), WithLogger(quiet()))
	if st := b3.Status(); st.DailyTokens != 0 || st.MonthlyTokens != 200 {
		t.Fatalf("restart on the next day: %+v, want day 0 and month 200", st)
	}
}

func TestAStoreThatFailsToSaveDoesNotStopTheGame(t *testing.T) {
	store := &memStore{failSave: true}
	b := New(Limits{DailyTokens: 100}, store, WithLogger(quiet()))
	c := b.Wrap(usageClient(60, 0))
	if err := call(t, c); err != nil {
		t.Fatal(err)
	}
	_ = call(t, c)
	if err := call(t, c); !errors.Is(err, model.ErrBudgetExceeded) {
		t.Fatalf("the in-memory counter must still cap when persisting fails: %v", err)
	}
}

func TestSetEnabledReportsAStoreFailure(t *testing.T) {
	b := New(DefaultLimits(), failingSwitch{}, WithLogger(quiet()))
	if err := b.SetEnabled(false); err == nil {
		t.Fatal("want an error, so the admin is not told the switch is off when it is not")
	}
	if st := b.Status(); !st.Enabled {
		t.Fatal("a switch that did not persist must not change behaviour")
	}
}

type failingSwitch struct{}

func (failingSwitch) Load(context.Context, string, string) (Snapshot, error) {
	return Snapshot{}, nil
}
func (failingSwitch) SaveTokens(context.Context, string, int64, string, int64) error { return nil }
func (failingSwitch) SaveDisabled(context.Context, bool) error {
	return errors.New("read-only")
}

func TestWrapKeepsANilClientNil(t *testing.T) {
	if New(DefaultLimits(), nil).Wrap(nil) != nil {
		t.Fatal("no transport must stay no transport, or a keyless server would offer model tiers")
	}
}

func TestLimitsFromEnv(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	got, err := limitsFrom(env(nil))
	if err != nil || got != DefaultLimits() {
		t.Fatalf("unset: %+v, %v, want the defaults (caps ON)", got, err)
	}
	got, err = limitsFrom(env(map[string]string{EnvDailyTokens: "500", EnvMonthlyTokens: "0"}))
	if err != nil || got.DailyTokens != 500 || got.MonthlyTokens != 0 {
		t.Fatalf("set: %+v, %v", got, err)
	}
	for _, bad := range []string{"lots", "-1", "1.5"} {
		if _, err := limitsFrom(env(map[string]string{EnvDailyTokens: bad})); err == nil {
			t.Errorf("%q: want a boot error", bad)
		}
	}
}

// --- the policy end to end -------------------------------------------

// pinnedB is a Layer B that always answers index 0 (the pass).
type pinnedB struct{ calls int }

func (p *pinnedB) Name() string { return "pinned-b" }
func (p *pinnedB) Decide(context.Context, aiseat.Input) (aiseat.Decision, error) {
	p.calls++
	return aiseat.Decision{Index: 0, Reason: "heuristic"}, nil
}

func castWindow() aiseat.Input {
	me := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	opp := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mv := func(k legal.Kind, label, params string) legal.Move {
		m := legal.Move{Kind: k, Label: label, Player: me}
		if params != "" {
			m.Params = []byte(params)
		}
		return m
	}
	return aiseat.Input{
		Seat: me,
		View: protocol.GameView{
			Seats: []protocol.PlayerView{
				{ID: me.String(), Name: "Bot0", Seat: 0, Life: 40,
					Hand:    protocol.ZoneView{Kind: "hand", Count: 1, Cards: []protocol.CardView{{InstanceID: "bear", Name: "Bear", ManaCost: "{1}{R}", TypeLine: "Creature — Bear"}}},
					Library: protocol.ZoneView{Kind: "library", Count: 60}},
				{ID: opp.String(), Name: "Bot1", Seat: 1, Life: 31,
					Hand:    protocol.ZoneView{Kind: "hand", Count: 4},
					Library: protocol.ZoneView{Kind: "library", Count: 55}},
			},
			Turn: protocol.TurnView{Seq: 7, Number: 7, ActiveSeat: 0, PriorityHolder: 0, Step: "precombat_main"},
		},
		Moves: []legal.Move{
			mv(legal.KindPass, "Pass priority", ""),
			mv(legal.KindCast, "Cast Bear", `{"instance_id":"bear"}`),
			mv(legal.KindCast, "Cast Bear again", `{"instance_id":"bear"}`),
		},
	}
}

func decide(t *testing.T, p *model.Policy) aiseat.Decision {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	d, err := p.Decide(ctx, castWindow())
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	return d
}

// The headline behaviour: a model seat whose server hit its cap plays
// exactly the heuristic's move, makes no model call, counts the window
// as budget-capped rather than as an error, and never claims it
// attempted a call.
func TestCappedModelSeatPlaysTheHeuristicWithNoCallAndNoError(t *testing.T) {
	inner := model.AlwaysIndex(1)
	inner.Reply = func(int, model.Request) (model.Response, error) {
		return model.Response{Text: `{"index": 1, "why": "fake"}`, Usage: model.Usage{InputTokens: 80, OutputTokens: 20}}, nil
	}
	b := New(Limits{DailyTokens: 150}, nil, WithLogger(quiet()))
	h := &pinnedB{}
	cfg := model.DefaultConfig()
	cfg.Fallback = h
	cfg.Client = b.Wrap(inner)
	cfg.Log = quiet()
	p := model.New(cfg)

	// Two decisions fit (100, then 200 crosses the 150 line).
	for i := 0; i < 2; i++ {
		if d := decide(t, p); d.Index != 1 {
			t.Fatalf("decision %d: index = %d, want the model's 1", i, d.Index)
		}
	}
	callsBefore := inner.Calls()

	d := decide(t, p)
	if d.Index != 0 {
		t.Fatalf("over the cap: index = %d, want the heuristic's 0", d.Index)
	}
	if inner.Calls() != callsBefore {
		t.Fatalf("a call left the process over the cap (%d -> %d)", callsBefore, inner.Calls())
	}
	st := p.Stats()
	if st.ByFallback[model.FallbackBudgetCapped] != 1 {
		t.Fatalf("ByFallback = %v, want one %q", st.ByFallback, model.FallbackBudgetCapped)
	}
	if st.ByFallback[model.FallbackError] != 0 {
		t.Fatalf("ByFallback = %v: a cap is not an outage", st.ByFallback)
	}
	if st.ModelCalls != 2 {
		t.Fatalf("ModelCalls = %d, want 2: a refused call is not a call", st.ModelCalls)
	}
}

func TestKillSwitchStopsARunningModelSeat(t *testing.T) {
	b := New(DefaultLimits(), nil, WithLogger(quiet()))
	cfg := model.DefaultConfig()
	cfg.Fallback = &pinnedB{}
	cfg.Client = b.Wrap(model.AlwaysIndex(1))
	cfg.Log = quiet()
	p := model.New(cfg)

	if d := decide(t, p); d.Index != 1 {
		t.Fatalf("before the switch: index = %d, want 1", d.Index)
	}
	if err := b.SetEnabled(false); err != nil {
		t.Fatal(err)
	}
	if d := decide(t, p); d.Index != 0 {
		t.Fatalf("after the switch: index = %d, want the heuristic's 0 on the very next decision", d.Index)
	}
	if got := p.Stats().ByFallback[model.FallbackBudgetCapped]; got != 1 {
		t.Fatalf("ByFallback[budget-capped] = %d, want 1", got)
	}
}
