package aiseat_test

import (
	"context"
	"errors"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics/metricstest"
)

// totals_test.go is ADR 0123 §11's check on the runner half of §3's
// bot totals: a runner's decisions, decision times and fallbacks reach
// the process-wide cmdctrl_bot_* families exactly as they reach its own
// Stats, labelled by tier and by the layer that answered. Stepped
// runners, so a step is the whole schedule and nothing runs after it.

func bot(t *testing.T, name string, labels metricstest.L) float64 {
	t.Helper()
	return metricstest.Value(t, name, labels)
}

func stepOnce(t *testing.T, policy aiseat.Policy, maxThink time.Duration) *aiseat.Runner {
	t.Helper()
	room := newRoom(t, 2, 11)
	r := aiseat.NewStepped(room, room.Game.Seats[0].ID, policy, aiseat.Config{MaxThink: maxThink}, testLogger())
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	r.Step(ctx)
	return r
}

// The random tier: every decision is layer "random", and every window
// is timed.
func TestRandomSeatDecisionsReachTheTotals(t *testing.T) {
	tier := metrics.BotTierRandom
	decisions := bot(t, "cmdctrl_bot_decisions_total", metricstest.L{"tier": tier})
	random := bot(t, "cmdctrl_bot_decisions_total", metricstest.L{"tier": tier, "layer": "random"})
	timed := bot(t, "cmdctrl_bot_decision_seconds", metricstest.L{"tier": tier})

	r := stepOnce(t, aiseat.NewRandomPolicy(rand.NewPCG(1, 2)), 2*time.Second)
	st := r.Stats()
	if st.Decisions == 0 {
		t.Fatal("the seat made no decision in its first step")
	}
	if got := bot(t, "cmdctrl_bot_decisions_total", metricstest.L{"tier": tier}) - decisions; got != float64(st.Decisions) {
		t.Errorf("decisions rose by %v, the runner counted %d", got, st.Decisions)
	}
	if got := bot(t, "cmdctrl_bot_decisions_total", metricstest.L{"tier": tier, "layer": "random"}) - random; got != float64(st.Decisions) {
		t.Errorf("layer=random rose by %v, want all %d", got, st.Decisions)
	}
	if got := bot(t, "cmdctrl_bot_decision_seconds", metricstest.L{"tier": tier}) - timed; got < float64(st.Decisions) {
		t.Errorf("decision seconds observed %v windows, fewer than %d decisions", got, st.Decisions)
	}
}

// A policy that always fails: every window is a runner fallback with
// cause policy-error, and none is a decision. Its name is no tier, so
// it is "other".
func TestRunnerFallbacksReachTheTotals(t *testing.T) {
	tier := metrics.BotTierOther
	cause := bot(t, "cmdctrl_bot_fallbacks_total", metricstest.L{"tier": tier, "cause": "policy-error"})
	decisions := bot(t, "cmdctrl_bot_decisions_total", metricstest.L{"tier": tier})

	r := stepOnce(t, failing{err: errors.New("boom")}, 2*time.Second)
	st := r.Stats()
	if st.Fallbacks == 0 {
		t.Fatal("no fallback in the first step")
	}
	if got := bot(t, "cmdctrl_bot_fallbacks_total", metricstest.L{"tier": tier, "cause": "policy-error"}) - cause; got != float64(st.Fallbacks) {
		t.Errorf("policy-error fallbacks rose by %v, the runner counted %d", got, st.Fallbacks)
	}
	if got := bot(t, "cmdctrl_bot_decisions_total", metricstest.L{"tier": tier}) - decisions; got != 0 {
		t.Errorf("decisions rose by %v for a policy that never decided", got)
	}
}

// The heuristic tier: Layer A's windows are noted A by the rules
// filter, the rest are the heuristic's, B, and the two add up to the
// runner's decisions; the filter's meter says how many were A.
func TestHeuristicSeatLayersReachTheTotals(t *testing.T) {
	tier := metrics.BotTierHeuristic
	a := bot(t, "cmdctrl_bot_decisions_total", metricstest.L{"tier": tier, "layer": "A"})
	b := bot(t, "cmdctrl_bot_decisions_total", metricstest.L{"tier": tier, "layer": "B"})

	meter := &rules.Meter{}
	f := rules.NewFilter(heuristic.New(), meter)
	f.Tier = string(tiers.Heuristic)
	r := stepOnce(t, f, 2*time.Second)
	st := r.Stats()
	if st.Decisions == 0 {
		t.Fatal("the seat made no decision in its first step")
	}
	gotA := bot(t, "cmdctrl_bot_decisions_total", metricstest.L{"tier": tier, "layer": "A"}) - a
	gotB := bot(t, "cmdctrl_bot_decisions_total", metricstest.L{"tier": tier, "layer": "B"}) - b
	if gotA+gotB != float64(st.Decisions) {
		t.Errorf("A %v + B %v != the runner's %d decisions", gotA, gotB, st.Decisions)
	}
	if ms := meter.Stats(); gotA != float64(ms.Absorbed) {
		t.Errorf("layer A rose by %v, the meter absorbed %d", gotA, ms.Absorbed)
	}
}

// A model tier on a FakeClient: the runner's per-layer decisions match
// the funnel's own per-layer count, and its model calls match the
// funnel's, all with no network.
func TestModelSeatLayersAndCallsReachTheTotals(t *testing.T) {
	tier := metrics.BotTierAssisted
	layers := []string{"A", "B", "C"}
	before := map[string]float64{}
	for _, l := range layers {
		before[l] = bot(t, "cmdctrl_bot_decisions_total", metricstest.L{"tier": tier, "layer": l})
	}
	calls := bot(t, "cmdctrl_bot_model_calls_total", metricstest.L{"tier": tier})

	cfg := model.DefaultConfig()
	cfg.Client = model.EchoesFallback()
	cfg.Fallback = heuristic.New()
	cfg.Log = testLogger()
	cfg.Improvise = false
	p := model.New(cfg)
	r := stepOnce(t, p, 5*time.Second)

	st, ps := r.Stats(), p.Stats()
	if st.Decisions == 0 {
		t.Fatal("the seat made no decision in its first step")
	}
	if st.Fallbacks != 0 {
		t.Fatalf("the runner fell back %d times; the per-layer comparison needs every window decided", st.Fallbacks)
	}
	var sum float64
	for _, l := range layers {
		got := bot(t, "cmdctrl_bot_decisions_total", metricstest.L{"tier": tier, "layer": l}) - before[l]
		sum += got
		if got != float64(ps.ByLayer[l]) {
			t.Errorf("layer %s rose by %v, the funnel counted %d", l, got, ps.ByLayer[l])
		}
	}
	if sum != float64(st.Decisions) {
		t.Errorf("decisions rose by %v, the runner counted %d", sum, st.Decisions)
	}
	if got := bot(t, "cmdctrl_bot_model_calls_total", metricstest.L{"tier": tier}) - calls; got != float64(ps.ModelCalls) {
		t.Errorf("model calls rose by %v, the funnel made %d", got, ps.ModelCalls)
	}
	if err := metrics.CheckClosedLabels(metrics.Registry); err != nil {
		t.Error(err)
	}
}

// Every runner fallback cause is its own value of the cause label;
// none falls to "other".
func TestEveryRunnerFallbackCauseIsALabel(t *testing.T) {
	for _, c := range []string{
		aiseat.FallbackTimeout, aiseat.FallbackPolicyError, aiseat.FallbackOutOfRange,
		aiseat.FallbackDeclinePass, aiseat.FallbackDeclineAlwaysLegal,
	} {
		before := bot(t, "cmdctrl_bot_fallbacks_total", metricstest.L{"tier": "other", "cause": c})
		metrics.RecordRunnerFallback("cause-check", c)
		if got := bot(t, "cmdctrl_bot_fallbacks_total", metricstest.L{"tier": "other", "cause": c}) - before; got != 1 {
			t.Errorf("runner cause %q did not land on its own label (rose by %v)", c, got)
		}
	}
}

// The tier label's closed set is aiseat's four tiers: each is its own
// value, nothing falls to "other".
func TestEveryTierIsALabel(t *testing.T) {
	for _, tr := range tiers.All() {
		if got := metrics.BotTierLabel(string(tr)); got != string(tr) {
			t.Errorf("tier %q is labelled %q", tr, got)
		}
	}
}
