package model

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// funnel_test.go exercises the layer routing and every failure mode
// the model can produce. It uses a stub Layer B rather than the real
// heuristic so that "what the fallback would have done" is a fixed
// number, and a FakeClient rather than a network so the failures are
// the ones being tested rather than the weather.

// --- fixtures -------------------------------------------------------

var (
	meSeat  = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	oppSeat = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

// stubB is a pinned Layer B: it always answers Index and it exposes a
// ranking when one is set.
type stubB struct {
	index    int
	reason   string
	rank     []heuristic.Candidate
	err      error
	concedes bool
	calls    int
}

func (s *stubB) Name() string { return "stub-b" }

func (s *stubB) Decide(context.Context, aiseat.Input) (aiseat.Decision, error) {
	s.calls++
	if s.err != nil {
		return aiseat.Decision{}, s.err
	}
	return aiseat.Decision{Index: s.index, Reason: s.reason}, nil
}

func (s *stubB) Rank(context.Context, aiseat.Input) []heuristic.Candidate { return s.rank }

func (s *stubB) ShouldConcede(aiseat.Input) bool { return s.concedes }

func mv(kind legal.Kind, label string, params string) legal.Move {
	m := legal.Move{Kind: kind, Label: label, Player: meSeat}
	if params != "" {
		m.Params = []byte(params)
	}
	return m
}

func pass() legal.Move { return mv(legal.KindPass, "Pass priority", "") }

// castWindow is a window a model is allowed to think about: a pass
// and two castable spells, so Layer A escalates.
func castWindow() aiseat.Input {
	return aiseat.Input{
		Seat: meSeat,
		View: view(),
		Moves: []legal.Move{
			pass(),
			mv(legal.KindCast, "Cast Lightning Bolt targeting Opponent", `{"instance_id":"bolt","targets":[{"kind":"player","id":"`+oppSeat.String()+`"}]}`),
			mv(legal.KindCast, "Cast Bear", `{"instance_id":"bear"}`),
		},
	}
}

func view() protocol.GameView {
	return protocol.GameView{
		Seats: []protocol.PlayerView{
			{
				ID: meSeat.String(), Name: "Bot0", Seat: 0, Life: 40,
				Hand: protocol.ZoneView{Kind: "hand", Count: 2, Cards: []protocol.CardView{
					{InstanceID: "bolt", Name: "Lightning Bolt", ManaCost: "{R}", TypeLine: "Instant"},
					{InstanceID: "bear", Name: "Bear", ManaCost: "{1}{R}", TypeLine: "Creature — Bear"},
				}},
				Library: protocol.ZoneView{Kind: "library", Count: 60},
			},
			{
				ID: oppSeat.String(), Name: "Bot1", Seat: 1, Life: 31,
				Hand:    protocol.ZoneView{Kind: "hand", Count: 4},
				Library: protocol.ZoneView{Kind: "library", Count: 55},
			},
		},
		Turn: protocol.TurnView{Number: 7, ActiveSeat: 0, PriorityHolder: 0, Step: "precombat_main"},
	}
}

// testPolicy builds a funnel with a pinned Layer B and no logging
// noise.
func testPolicy(t *testing.T, client Client, b *stubB, tweak func(*Config)) *Policy {
	t.Helper()
	cfg := DefaultConfig()
	cfg.Fallback = b
	cfg.Client = client
	cfg.Log = testLogger()
	if tweak != nil {
		tweak(&cfg)
	}
	return New(cfg)
}

func decide(t *testing.T, p *Policy, in aiseat.Input, budget time.Duration) aiseat.Decision {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), budget)
	defer cancel()
	d, err := p.Decide(ctx, in)
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	return d
}

// --- layer routing --------------------------------------------------

// Layer A's whole reason to exist: a window it settles costs nothing,
// which means it never touches the model AND never touches Layer B.
func TestAbsorbedWindowsNeverReachTheModelOrTheHeuristic(t *testing.T) {
	b := &stubB{index: 0, reason: "stub"}
	fake := AlwaysIndex(1)
	p := testPolicy(t, fake, b, nil)

	in := aiseat.Input{Seat: meSeat, View: view(), Moves: []legal.Move{
		pass(), mv(legal.KindMana, "Tap Mountain for R", ""),
	}}
	d := decide(t, p, in, 2*time.Second)
	if d.Index != 0 {
		t.Errorf("index = %d, want the pass", d.Index)
	}
	if fake.Calls() != 0 {
		t.Errorf("the model was called %d times on a window Layer A settles", fake.Calls())
	}
	if b.calls != 0 {
		t.Errorf("the heuristic was called %d times on a window Layer A settles", b.calls)
	}
	st := p.Stats()
	if st.ByLayer[LayerA] != 1 || st.Windows != 1 {
		t.Errorf("stats = %+v", st)
	}
	if st.AbsorptionRate() != 1 {
		t.Errorf("absorption rate = %v, want 1", st.AbsorptionRate())
	}
}

// A routine window gets the cheap model. This is the entire cost
// argument and it is one line of routing, so it gets a test.
func TestRoutineWindowAsksTheCheapModel(t *testing.T) {
	b := &stubB{index: 0, reason: "stub"}
	fake := AlwaysIndex(1)
	p := testPolicy(t, fake, b, func(c *Config) {
		// Neutralise every escalation trigger so the window is
		// genuinely routine.
		c.ThreatPower, c.ThreatCreature, c.DangerLife = 999, 999, 0
		c.Epsilon = 0
	})
	in := aiseat.Input{Seat: meSeat, View: view(), Moves: []legal.Move{
		pass(), mv(legal.KindCast, "Cast Bear", `{"instance_id":"bear"}`),
	}}
	d := decide(t, p, in, 2*time.Second)
	if d.Index != 1 {
		t.Fatalf("index = %d, want the model's 1", d.Index)
	}
	reqs := fake.Requests()
	if len(reqs) != 1 {
		t.Fatalf("calls = %d, want 1", len(reqs))
	}
	if reqs[0].Model != DefaultConfig().Routine.ID {
		t.Errorf("model = %q, want the routine model %q", reqs[0].Model, DefaultConfig().Routine.ID)
	}
	if st := p.Stats(); st.ByLayer[LayerC] != 1 {
		t.Errorf("stats = %+v", st)
	}
}

// An escalation trigger swaps in the frontier model.
func TestEscalationAsksTheFrontierModel(t *testing.T) {
	b := &stubB{index: 0, reason: "stub"}
	fake := AlwaysIndex(0)
	p := testPolicy(t, fake, b, nil)

	// A blocks window: ADR 0033 lists declare-blockers as a trigger.
	in := aiseat.Input{Seat: meSeat, View: view(), Moves: []legal.Move{
		mv(legal.KindBlock, "Block Wurm with Bear", ""),
		mv(legal.KindBlock, "Block Wurm with Drake", ""),
	}}
	decide(t, p, in, 2*time.Second)
	reqs := fake.Requests()
	if len(reqs) != 1 {
		t.Fatalf("calls = %d, want 1", len(reqs))
	}
	if reqs[0].Model != DefaultConfig().Frontier.ID {
		t.Errorf("model = %q, want the frontier model %q", reqs[0].Model, DefaultConfig().Frontier.ID)
	}
	st := p.Stats()
	if st.ByEscalation[ReasonCombat] != 1 {
		t.Errorf("combat trigger did not fire: %+v", st.ByEscalation)
	}
}

// --- every way the model can fail ------------------------------------

// The table of failure modes the sub-PR spec names, each one landing
// on Layer B's answer with the reason recorded.
func TestEveryModelFailureFallsBackToTheHeuristic(t *testing.T) {
	cases := []struct {
		name     string
		client   Client
		fallback string
	}{
		{"total outage", AlwaysFails(ErrOutage), FallbackError},
		{"http error", AlwaysFails(&APIError{Status: 529, Kind: "overloaded_error"}), FallbackError},
		{"prose instead of json", AlwaysText("I think you should cast the Bear."), FallbackMalformed},
		{"empty reply", AlwaysText(""), FallbackMalformed},
		{"index past the end", AlwaysIndex(99), FallbackOutOfRange},
		{"negative index", AlwaysIndex(-4), FallbackOutOfRange},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := &stubB{index: 2, reason: "the heuristic's pick"}
			p := testPolicy(t, tc.client, b, nil)
			d := decide(t, p, castWindow(), 2*time.Second)
			if d.Index != 2 {
				t.Errorf("index = %d, want Layer B's 2", d.Index)
			}
			if d.Reason != "the heuristic's pick" {
				t.Errorf("reason = %q, want Layer B's", d.Reason)
			}
			st := p.Stats()
			if st.ByLayer[LayerB] != 1 {
				t.Errorf("the window was not attributed to Layer B: %+v", st.ByLayer)
			}
			if st.ByFallback[tc.fallback] != 1 {
				t.Errorf("fallback reasons = %v, want one %q", st.ByFallback, tc.fallback)
			}
		})
	}
}

// -1 is aiseat.Decline, not an out-of-range index: a policy may
// legitimately want none of the moves in a blocks window. Layer B
// owns that decision and Layer C must not corrupt it.
func TestDeclineFromLayerBSurvivesAModelFailure(t *testing.T) {
	b := &stubB{index: aiseat.Decline, reason: "nothing worth blocking"}
	p := testPolicy(t, AlwaysFails(ErrOutage), b, nil)
	in := aiseat.Input{Seat: meSeat, View: view(), Moves: []legal.Move{
		mv(legal.KindBlock, "Block Wurm with Bear", ""),
		mv(legal.KindBlock, "Block Wurm with Drake", ""),
	}}
	d := decide(t, p, in, 2*time.Second)
	if d.Index != aiseat.Decline {
		t.Errorf("index = %d, want aiseat.Decline", d.Index)
	}
}

// ADR 0033 §10: MaxThink is a hard deadline and the table never waits
// on a bot. A model that simply never answers must cost the budget
// and then hand over, not blow through it.
func TestASlowModelDegradesInsideTheDeadline(t *testing.T) {
	b := &stubB{index: 1, reason: "the heuristic's pick"}
	p := testPolicy(t, NeverAnswers(), b, func(c *Config) {
		c.Reserve = 200 * time.Millisecond
		c.MinBudget = 50 * time.Millisecond
	})
	const deadline = 600 * time.Millisecond
	started := time.Now()
	d := decide(t, p, castWindow(), deadline)
	elapsed := time.Since(started)

	if d.Index != 1 {
		t.Errorf("index = %d, want Layer B's 1", d.Index)
	}
	// The budget is deadline - Reserve, so the call should give up
	// with the reserve still unspent.
	if elapsed >= deadline-100*time.Millisecond {
		t.Errorf("Decide took %v of a %v deadline; the reserve was spent and the runner would have force-passed",
			elapsed, deadline)
	}
	if st := p.Stats(); st.ByFallback[FallbackError] != 1 {
		t.Errorf("fallbacks = %v", st.ByFallback)
	}
}

// With almost no time left there is no point dialling at all.
func TestNoBudgetMeansNoCall(t *testing.T) {
	b := &stubB{index: 1, reason: "stub"}
	fake := AlwaysIndex(2)
	p := testPolicy(t, fake, b, func(c *Config) {
		c.Reserve = 250 * time.Millisecond
		c.MinBudget = 200 * time.Millisecond
	})
	d := decide(t, p, castWindow(), 300*time.Millisecond)
	if fake.Calls() != 0 {
		t.Errorf("the model was dialled with %v of budget", 300*time.Millisecond-250*time.Millisecond)
	}
	if d.Index != 1 {
		t.Errorf("index = %d, want Layer B's", d.Index)
	}
	if st := p.Stats(); st.ByFallback[FallbackNoBudget] != 1 {
		t.Errorf("fallbacks = %v", st.ByFallback)
	}
}

// A deployment with no API key is a supported deployment: the tier
// keeps its name and plays on Layer A + B.
func TestNilClientIsACompletePolicy(t *testing.T) {
	b := &stubB{index: 1, reason: "stub"}
	p := testPolicy(t, nil, b, nil)
	d := decide(t, p, castWindow(), 2*time.Second)
	if d.Index != 1 {
		t.Errorf("index = %d, want Layer B's", d.Index)
	}
	if p.Name() != TierAssisted {
		t.Errorf("name = %q, want %q", p.Name(), TierAssisted)
	}
	st := p.Stats()
	if st.ByFallback[FallbackNoClient] != 1 || st.ModelCalls != 0 {
		t.Errorf("stats = %+v", st)
	}
}

// Layer B failing is the one case with nothing underneath. The error
// has to reach the runner, which takes the pass.
func TestALayerBErrorReachesTheRunner(t *testing.T) {
	b := &stubB{err: errors.New("scorer exploded")}
	p := testPolicy(t, AlwaysIndex(0), b, nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := p.Decide(ctx, castWindow()); err == nil {
		t.Fatal("a Layer B failure was swallowed; the runner never learns it needs its own fallback")
	}
	if st := p.Stats(); st.ByFallback[FallbackPolicyError] != 1 {
		t.Errorf("fallbacks = %v", st.ByFallback)
	}
}

// An empty move list is the Policy contract's error, not a model call.
func TestNoMovesIsAnError(t *testing.T) {
	p := testPolicy(t, AlwaysIndex(0), &stubB{}, nil)
	if _, err := p.Decide(context.Background(), aiseat.Input{Seat: meSeat}); !errors.Is(err, aiseat.ErrNoMoves) {
		t.Errorf("err = %v, want aiseat.ErrNoMoves", err)
	}
}

// --- the model's answer is honoured ----------------------------------

func TestAnInRangeIndexIsTaken(t *testing.T) {
	b := &stubB{index: 0, reason: "pass"}
	p := testPolicy(t, AlwaysIndex(2), b, nil)
	d := decide(t, p, castWindow(), 2*time.Second)
	if d.Index != 2 {
		t.Fatalf("index = %d, want the model's 2", d.Index)
	}
	if !strings.Contains(d.Reason, "fake") {
		t.Errorf("reason = %q, want the model's own words", d.Reason)
	}
	recs := p.Records()
	if len(recs) != 1 || recs[0].Layer != LayerC {
		t.Fatalf("records = %+v", recs)
	}
	if !recs[0].Attempted || recs[0].Model == "" {
		t.Errorf("record does not say a call was made: %+v", recs[0])
	}
}

func TestConcedeForwardsToLayerB(t *testing.T) {
	b := &stubB{concedes: true}
	p := testPolicy(t, nil, b, nil)
	if !p.ShouldConcede(castWindow()) {
		t.Error("the funnel swallowed Layer B's concede")
	}
	b.concedes = false
	if p.ShouldConcede(castWindow()) {
		t.Error("the funnel invented a concede")
	}
}
