package aiseat_test

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// model_game_test.go runs Layer C against the real engine.
//
// Every test here uses a FakeClient, because this repository has no
// model endpoint in CI and no API key in the environment. That is a
// real limit and it is worth naming precisely rather than leaving to
// be discovered: these tests prove the PLUMBING — prompt assembly,
// model selection, the deadline arithmetic, index validation, and the
// fallback under every failure — and they prove nothing at all about
// whether a model plays Magic well. The first is what can be verified
// offline; the second needs a key and a game.
//
// The harness is heuristic_game_test.go's: playGame seats one policy
// per player, runs to a winner, and fails the test on a stall.

// quietLogger discards. Several tests here deliberately break the
// model on every window, and the policy is right to warn about each
// one — but a few thousand warnings is not a test log anybody reads,
// and the assertions are on Stats rather than on what was printed.
func quietLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// modelSeats builds n `assisted` seats sharing a meter and a client.
func modelSeats(t *testing.T, n int, client model.Client, meter *rules.Meter, tweak func(*model.Config)) ([]aiseat.Policy, []*model.Policy) {
	t.Helper()
	policies := make([]aiseat.Policy, 0, n)
	funnels := make([]*model.Policy, 0, n)
	for i := 0; i < n; i++ {
		cfg := model.DefaultConfig()
		cfg.Client = client
		cfg.Meter = meter
		cfg.Log = quietLogger()
		cfg.Deck = model.DeckProfile{
			Name:      "Mono-red battle deck (test harness)",
			Archetype: "Curve out, swing, point the burn at whoever is ahead.",
			Cards: []model.DeckCard{
				{Name: "Lightning Bolt", Cost: "{R}", Type: "Instant", Oracle: "Lightning Bolt deals 3 damage to any target."},
				{Name: "Mountain", Type: "Basic Land — Mountain", Oracle: "{T}: Add {R}."},
			},
		}
		if tweak != nil {
			tweak(&cfg)
		}
		p := model.New(cfg)
		policies = append(policies, p)
		funnels = append(funnels, p)
	}
	return policies, funnels
}

// totalModelStats folds the per-seat instrumentation into one table's
// worth of numbers.
func totalModelStats(funnels []*model.Policy) model.Stats {
	out := model.Stats{
		ByLayer:      map[string]int64{},
		ByEscalation: map[string]int64{},
		ByFallback:   map[string]int64{},
	}
	for _, f := range funnels {
		s := f.Stats()
		out.Windows += s.Windows
		out.ModelCalls += s.ModelCalls
		out.ModelLatency += s.ModelLatency
		if s.MaxModelLatency > out.MaxModelLatency {
			out.MaxModelLatency = s.MaxModelLatency
		}
		out.Usage.InputTokens += s.Usage.InputTokens
		out.Usage.OutputTokens += s.Usage.OutputTokens
		out.Usage.CacheReadTokens += s.Usage.CacheReadTokens
		out.Usage.CacheWriteTokens += s.Usage.CacheWriteTokens
		for k, v := range s.ByLayer {
			out.ByLayer[k] += v
		}
		for k, v := range s.ByEscalation {
			out.ByEscalation[k] += v
		}
		for k, v := range s.ByFallback {
			out.ByFallback[k] += v
		}
	}
	return out
}

func logModelStats(t *testing.T, label string, st model.Stats) {
	t.Helper()
	t.Logf("%s: %d windows — A %d, B %d, C %d (Layer A absorbed %.1f%%)",
		label, st.Windows, st.ByLayer[model.LayerA], st.ByLayer[model.LayerB], st.ByLayer[model.LayerC],
		st.AbsorptionRate()*100)
	t.Logf("  %d model calls, %d of them escalated to the frontier model (%.1f%% of surviving windows); worst call %v",
		st.ModelCalls, st.Escalated, st.EscalationRate()*100, st.MaxModelLatency)
	t.Logf("  reported tokens: in %d (cache read %d, cache write %d), out %d",
		st.Usage.InputTokens, st.Usage.CacheReadTokens, st.Usage.CacheWriteTokens, st.Usage.OutputTokens)
	for _, tr := range st.Triggers() {
		t.Logf("  escalation %-20s %d", tr, st.ByEscalation[tr])
	}
	for _, k := range []string{
		model.FallbackNoClient, model.FallbackNoBudget, model.FallbackError,
		model.FallbackMalformed, model.FallbackOutOfRange, model.FallbackPolicyError,
	} {
		if n := st.ByFallback[k]; n > 0 {
			t.Logf("  fallback   %-20s %d", k, n)
		}
	}
}

// The whole Layer C path against the real engine: prompt assembly,
// the call, the parse, the index bounds check and the dispatch, for a
// full four-player game.
func TestAssistedTierPlaysAWholeGameThroughTheModelPath(t *testing.T) {
	requireGameTests(t)
	const (
		turnBudget = 35
		wall       = 120 * time.Second
	)
	var meter rules.Meter
	client := model.EchoesFallback()
	policies, funnels := modelSeats(t, 4, client, &meter, nil)

	res := playGame(t, 3300, policies, turnBudget, wall)
	assertNoEnumeratorBugs(t, res)
	if res.state != game.StateEnded {
		t.Errorf("game did not finish inside %d turns (state %s, lives %v)", turnBudget, res.state, res.lives)
	}

	st := totalModelStats(funnels)
	logModelStats(t, "assisted (fake client)", st)
	if st.ByLayer[model.LayerC] == 0 {
		t.Fatal("no window reached Layer C; the model path never ran")
	}
	if st.ModelCalls != int64(client.Calls()) {
		t.Errorf("the policy counted %d model calls, the client saw %d", st.ModelCalls, client.Calls())
	}
	if st.ByFallback[model.FallbackOutOfRange] != 0 || st.ByFallback[model.FallbackMalformed] != 0 {
		t.Errorf("the fake produced unusable answers: %v", st.ByFallback)
	}
	// The static half is what prompt caching pays for; it has to be
	// identical on every call of the game, and this is the only place
	// that check runs against a real sequence of boards.
	reqs := client.Requests()
	if len(reqs) < 2 {
		t.Fatalf("only %d requests", len(reqs))
	}
	for i, r := range reqs[1:] {
		if len(r.System) != len(reqs[0].System) {
			t.Fatalf("request %d has a different number of system blocks", i+1)
		}
		for j := range r.System {
			if r.System[j].Text != reqs[0].System[j].Text {
				t.Fatalf("request %d system block %d differs from the first — the prompt cache is dead", i+1, j)
			}
		}
	}
}

// The S31 exit criterion, verbatim: "Model-outage drill: Layer C
// hard-fails, game completes on Layer B, no frozen table."
//
// The endpoint answers for a while and then goes away for good, which
// is the outage that actually happens: the game started fine and the
// provider fell over in the middle of it.
func TestModelOutageDrill(t *testing.T) {
	requireGameTests(t)
	const (
		turnBudget = 35
		wall       = 120 * time.Second
		goodCalls  = 25
	)
	var meter rules.Meter
	client := model.FailsAfter(goodCalls, model.EchoesFallback(), model.ErrOutage)
	policies, funnels := modelSeats(t, 4, client, &meter, nil)

	res := playGame(t, 3400, policies, turnBudget, wall)

	// 1. The table did not freeze. playGame fails on a stall before it
	//    ever reaches here, so getting this far is half the criterion;
	//    the other half is that the game actually ENDED.
	if res.state != game.StateEnded {
		t.Errorf("the game did not finish after the model went away (state %s, lives %v)", res.state, res.lives)
	}
	if res.winner < 0 {
		t.Errorf("no single survivor: lives %v", res.lives)
	}

	// 2. Nothing illegal reached the engine, and — the load-bearing
	//    assertion — the RUNNER never had to use its own fallback.
	//    That fallback is the last resort that force-passes; if
	//    Layer B is doing its job it is never reached.
	assertNoEnumeratorBugs(t, res)

	st := totalModelStats(funnels)
	logModelStats(t, "outage drill", st)

	// 3. The outage is visible in the instrumentation rather than
	//    inferred: calls failed, and every failed window landed on
	//    Layer B.
	if st.ByFallback[model.FallbackError] == 0 {
		t.Fatal("no call failed; the drill did not drill anything")
	}
	if st.ByLayer[model.LayerC] > goodCalls {
		t.Errorf("Layer C answered %d windows but only %d calls could have succeeded",
			st.ByLayer[model.LayerC], goodCalls)
	}
	if st.ByLayer[model.LayerA]+st.ByLayer[model.LayerB]+st.ByLayer[model.LayerC] != st.Windows {
		t.Errorf("windows do not add up: %+v", st.ByLayer)
	}
	// And the game was still overwhelmingly decided by Layer A and
	// Layer B, which is the point of building them.
	if free := st.ByLayer[model.LayerA] + st.ByLayer[model.LayerB]; free < st.Windows*9/10 {
		t.Errorf("only %d of %d windows were answered without a usable model", free, st.Windows)
	}
}

// **What is deliberately NOT a whole game here.** A model that
// replies with prose, or with an index that is not a move, or that is
// simply absent, all fall back to Layer B per window — and that is
// unit-tested exhaustively next door in aiseat/model
// (TestEveryModelFailureFallsBackToTheHeuristic). The only thing a
// whole game adds over the unit is "and the table still finishes",
// which the outage drill above already demonstrates on the same code
// path. This package runs under `go test -race -cover` in CI against
// a ten-minute per-package cap and already spends most of it, so the
// games here are rationed to the ones that show something a unit
// cannot.

// ADR 0033 §10: "A bot that cannot decide passes. The table never
// waits on a model." A model that never answers must cost its budget
// and hand over to Layer B — not blow MaxThink and leave the runner
// to force-pass, which is a bot that stops playing rather than a bot
// that plays worse.
func TestASlowModelDegradesRatherThanForcingPasses(t *testing.T) {
	requireGameTests(t)
	const (
		turnBudget = 20
		wall       = 120 * time.Second
	)
	var meter rules.Meter
	policies, funnels := modelSeats(t, 4, model.NeverAnswers(), &meter, func(c *model.Config) {
		// A realistic two-second timeout cannot be simulated at real
		// speed in a test — a whole game of them is an hour — so the
		// budget is scaled down and the ARITHMETIC is what is under
		// test: the call's deadline is always strictly inside the
		// runner's.
		c.MaxCall = 5 * time.Millisecond
		c.Reserve = 5 * time.Millisecond
		c.MinBudget = 1 * time.Millisecond
	})
	res := playGame(t, 3600, policies, turnBudget, wall)

	// The runner's own fallback counter is the assertion: it goes up
	// exactly when a policy misses the deadline or errors, and the
	// whole design of Layer C is that it cannot.
	total := res.totals()
	if total.Fallbacks != 0 {
		t.Errorf("the runner fell back %d times; a slow model blew MaxThink instead of degrading", total.Fallbacks)
	}
	assertNoEnumeratorBugs(t, res)
	st := totalModelStats(funnels)
	logModelStats(t, "slow model", st)
	if st.ByFallback[model.FallbackError] == 0 {
		t.Fatal("no call timed out; the test measured nothing")
	}
	if st.MaxModelLatency > 500*time.Millisecond {
		t.Errorf("a model call ran for %v against a 5ms budget", st.MaxModelLatency)
	}
}
