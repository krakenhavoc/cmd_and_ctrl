package aiseat_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// funnel_game_test.go measures ADR 0033 §5's funnel against the real
// engine. The sprint states the acceptance number rather than a
// preference — "if Layer A is not absorbing >80% of windows, the
// funnel is broken and should be fixed before reaching for a cheaper
// model" — so this is the test that either produces that number or
// fails the build.
//
// It reuses heuristic_game_test.go's harness: playGame seats one
// policy per player, runs to a winner, and fails the test on a stall.
//
// **On the game budget.** This package already spends ~330s in CI
// before anything here runs, against a 10-minute per-package cap, so
// whole games are rationed: the two assertions below share one set of
// games rather than each taking their own, and the turn budget is the
// smallest that still finishes. `AISEAT_FUNNEL_GAMES=N` widens the
// sample when somebody is actually tuning.

// absorptionTarget is ADR 0033 §5's number, verbatim.
const absorptionTarget = 0.80

// Two assertions over one set of four-bot games:
//
//  1. Layer A absorbs more than 80% of priority windows.
//  2. Every window it absorbed, the heuristic underneath would have
//     answered identically — so the filter is a cost optimisation and
//     not a play regression hiding inside one.
//
// Every seat shares one Meter, so the rate is a property of the table
// rather than of a seat that happened to get a quiet game.
func TestLayerAAbsorbsMostWindowsAndAgreesWithTheHeuristic(t *testing.T) {
	games := 2
	if n, err := strconv.Atoi(os.Getenv("AISEAT_FUNNEL_GAMES")); err == nil && n > 0 {
		games = n
	}
	const (
		turnBudget = 40
		wall       = 120 * time.Second
	)
	var meter rules.Meter
	var checked int
	var disagreements []string

	for i := 0; i < games; i++ {
		seed := uint64(3100 + i)
		policies := make([]aiseat.Policy, 4)
		checkers := make([]*agreementPolicy, 4)
		for j := range policies {
			inner := heuristic.New()
			checkers[j] = &agreementPolicy{inner: inner, filter: rules.NewFilter(inner, &meter)}
			policies[j] = checkers[j]
		}
		res := playGame(t, seed, policies, turnBudget, wall)
		assertNoEnumeratorBugs(t, res)
		if res.state != game.StateEnded {
			t.Errorf("seed %d: game did not finish inside %d turns (state %s, lives %v)",
				seed, turnBudget, res.state, res.lives)
		}
		for _, c := range checkers {
			n, d := c.report()
			checked += n
			disagreements = append(disagreements, d...)
		}
	}

	st := meter.Stats()
	t.Logf("Layer A over %d games: %d/%d windows absorbed — %.1f%%",
		games, st.Absorbed, st.Windows, st.Rate()*100)
	for _, rule := range st.Rules() {
		t.Logf("  %-12s %6d (%.1f%%)", rule, st.ByRule[rule], float64(st.ByRule[rule])/float64(st.Windows)*100)
	}
	if st.Windows < 500 {
		t.Fatalf("only %d windows sampled; the rate is not a measurement", st.Windows)
	}
	if st.Rate() < absorptionTarget {
		t.Errorf("Layer A absorbed %.1f%% of windows, under ADR 0033 §5's %.0f%% floor — "+
			"the funnel is leaking routine decisions to the paid layers and that is what to fix first",
			st.Rate()*100, absorptionTarget*100)
	}

	t.Logf("Layer A verdicts cross-checked against the heuristic: %d absorbed windows, %d disagreements",
		checked, len(disagreements))
	if checked < 100 {
		t.Fatalf("only %d absorbed windows cross-checked; that half of the test measured nothing", checked)
	}
	for _, d := range disagreements {
		t.Error(d)
	}
}

// agreementPolicy drives the filter and, on every window the filter
// absorbed, asks the heuristic underneath what it would have done.
// The heuristic is only ever consulted AFTER the filter has returned,
// so the two never nest and the policy's own mutex is never taken
// twice.
type agreementPolicy struct {
	inner  *heuristic.Policy
	filter aiseat.Policy

	mu        sync.Mutex
	checked   int
	disagreed []string
}

func (p *agreementPolicy) Name() string { return "agreement" }

func (p *agreementPolicy) Decide(ctx context.Context, in aiseat.Input) (aiseat.Decision, error) {
	v := rules.Resolve(in)
	d, err := p.filter.Decide(ctx, in)
	if err != nil || !v.Absorbed() {
		return d, err
	}
	alt, aerr := p.inner.Decide(ctx, in)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.checked++
	if aerr == nil && alt.Index != v.Index {
		p.disagreed = append(p.disagreed, fmt.Sprintf(
			"Layer A took %q (rule %s) where the heuristic took %q (%s)",
			moveLabel(in, v.Index), v.Rule, moveLabel(in, alt.Index), alt.Reason))
	}
	return d, err
}

func (p *agreementPolicy) ShouldConcede(in aiseat.Input) bool { return p.inner.ShouldConcede(in) }

func (p *agreementPolicy) report() (int, []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.checked, append([]string(nil), p.disagreed...)
}

func moveLabel(in aiseat.Input, i int) string {
	if i < 0 || i >= len(in.Moves) {
		return fmt.Sprintf("<index %d>", i)
	}
	return in.Moves[i].Label
}
