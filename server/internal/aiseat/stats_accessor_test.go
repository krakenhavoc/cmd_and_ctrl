package aiseat_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
)

// stats_accessor_test.go is #505 part 2's Runner-side half: a Runner
// exposes the outermost model funnel's stats in its wrapper chain,
// and nothing when it has none. See stats_accessor.go.

// statsPolicy is a minimal Policy whose PolicyStats is fixed, so
// Runner.PolicyStats can be exercised without standing up a real
// model funnel.
type statsPolicy struct {
	stats aiseat.PolicyStats
}

func (statsPolicy) Name() string { return "stats-fake" }

func (statsPolicy) Decide(_ context.Context, in aiseat.Input) (aiseat.Decision, error) {
	if len(in.Moves) == 0 {
		return aiseat.Decision{}, aiseat.ErrNoMoves
	}
	return aiseat.Decision{Index: 0, Reason: "fake"}, nil
}

func (s statsPolicy) PolicyStats() aiseat.PolicyStats { return s.stats }

// wrapPolicy wraps another Policy without implementing anything of
// its own — the same shape rules.Filter and model.Policy declare
// (see capability.go) — so a lookup through it exercises the
// outermost-implementer rule rather than a bare type assertion on
// the policy the runner was started with.
type wrapPolicy struct {
	aiseat.Policy
}

func (w wrapPolicy) Unwrap() aiseat.Policy { return w.Policy }

func TestRunnerPolicyStatsAbsentForARandomSeat(t *testing.T) {
	room := newRoom(t, 2, 101)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := aiseat.Start(ctx, room, room.Game.Seats[0].ID, aiseat.NewRandomPolicy(nil), aiseat.Config{}, nil, testLogger())
	t.Cleanup(func() { cancel(); <-r.Done() })
	// The random tier has no model.Policy anywhere in its chain: it
	// IS the whole chain, and it does not implement PolicyStatser.
	if _, ok := r.PolicyStats(); ok {
		t.Error("a random seat reported PolicyStats; it has no model funnel")
	}
}

func TestRunnerPolicyStatsFromTheOutermostPolicy(t *testing.T) {
	room := newRoom(t, 2, 102)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	want := aiseat.PolicyStats{Windows: 7, ModelCalls: 3}
	r := aiseat.Start(ctx, room, room.Game.Seats[0].ID, statsPolicy{stats: want}, aiseat.Config{}, nil, testLogger())
	t.Cleanup(func() { cancel(); <-r.Done() })
	waitFor(t, "the seat to decide at least once", func() bool {
		s := r.Stats()
		return s.Decisions > 0 || s.Applied > 0 || s.Passes > 0
	})
	got, ok := r.PolicyStats()
	if !ok {
		t.Fatal("PolicyStats() ok = false, want true")
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("PolicyStats() = %+v, want %+v", got, want)
	}
}

// A policy that WRAPS a PolicyStatser still reports it — the same
// outermost-implementer rule Spender and every other optional
// extension gets from aiseat.Capability.
func TestRunnerPolicyStatsForwardsThroughAWrapper(t *testing.T) {
	room := newRoom(t, 2, 103)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	want := aiseat.PolicyStats{Windows: 1}
	inner := statsPolicy{stats: want}
	r := aiseat.Start(ctx, room, room.Game.Seats[0].ID, wrapPolicy{inner}, aiseat.Config{}, nil, testLogger())
	t.Cleanup(func() { cancel(); <-r.Done() })
	waitFor(t, "the seat to decide at least once", func() bool {
		s := r.Stats()
		return s.Decisions > 0 || s.Applied > 0 || s.Passes > 0
	})
	got, ok := r.PolicyStats()
	if !ok {
		t.Fatal("PolicyStats() did not forward through the wrapper")
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("PolicyStats() = %+v, want %+v", got, want)
	}
}
