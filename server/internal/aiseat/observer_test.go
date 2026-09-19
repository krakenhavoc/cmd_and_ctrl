package aiseat_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// observer_test.go covers the runner's half of the decision trace:
// that every window produces exactly one event, that the event says
// WHY the runner overruled the policy when it did, and that the
// latency ring behind Stats().Latency is actually filled.
//
// The classification is the part worth testing hardest. Before this,
// "the policy timed out" and "the policy chose the pass" were the
// same two log lines and the same Stats.Fallbacks counter; a report
// built on that could not tell a healthy heuristic table from an
// `assisted` table whose model missed every deadline.

type collector struct {
	mu  sync.Mutex
	evs []aiseat.DecisionEvent
}

func (c *collector) Observe(ev aiseat.DecisionEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evs = append(c.evs, ev)
}

func (c *collector) events() []aiseat.DecisionEvent {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]aiseat.DecisionEvent(nil), c.evs...)
}

func (c *collector) len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.evs)
}

// runWithObserver seats policy on seat 0 with an observer and a
// keep-and-pass opponent on seat 1, and runs until the observer has
// seen want events.
func runWithObserver(t *testing.T, policy aiseat.Policy, cfg aiseat.Config, want int) (*collector, *aiseat.Runner) {
	t.Helper()
	return runWithObserverUntil(t, policy, cfg, fmt.Sprintf("%d observer events", want),
		func(c *collector) bool { return c.len() >= want })
}

// runWithObserverUntil is the same table, waiting on a shape of event
// rather than on a count of them.
//
// The observed seat is seated FIRST, and that ordering is load-bearing
// now that aiseat.Start subscribes before it returns (#938): seat 1's
// keep therefore cannot land in a window where seat 0 is not listening,
// so a policy that parks on its first window is still woken for a
// second. Before that, seat 0 could step once against a board seat 1
// had already settled, park, and wait out the whole budget for a commit
// nobody was going to make.
func runWithObserverUntil(t *testing.T, policy aiseat.Policy, cfg aiseat.Config, what string, ready func(*collector) bool) (*collector, *aiseat.Runner) {
	t.Helper()
	room := newRoom(t, 2, 7)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	c := &collector{}
	cfg.Observer = c
	r := aiseat.Start(ctx, room, room.Game.Seats[0].ID, policy, cfg, nil, testLogger())
	aiseat.Start(ctx, room, room.Game.Seats[1].ID, &scripted{prefer: []string{"Keep hand"}}, aiseat.Config{}, nil, testLogger())
	waitFor(t, what, func() bool { return ready(c) })
	cancel()
	<-r.Done()
	return c, r
}

func TestRunnerObserverSeesEveryWindow(t *testing.T) {
	c, r := runWithObserver(t, heuristic.New(), aiseat.Config{}, 8)
	evs := c.events()
	st := r.Stats()

	applied := 0
	for i, ev := range evs {
		if len(ev.Input.Moves) == 0 {
			t.Fatalf("event %d has no move list; an observed window always had something to decide", i)
		}
		if ev.Seat != r.Seat() {
			t.Errorf("event %d is for seat %s, want %s", i, ev.Seat, r.Seat())
		}
		if ev.Policy != r.PolicyName() {
			t.Errorf("event %d names policy %q, want %q", i, ev.Policy, r.PolicyName())
		}
		if ev.Latency <= 0 {
			t.Errorf("event %d has no latency", i)
		}
		if ev.Applied {
			applied++
			if ev.Seq == 0 {
				t.Errorf("event %d was applied but carries no room seq", i)
			}
			if ev.Index < 0 || ev.Index >= len(ev.Input.Moves) {
				t.Errorf("event %d dispatched index %d of %d moves", i, ev.Index, len(ev.Input.Moves))
			}
			if ev.Label != ev.Input.Moves[ev.Index].Label {
				t.Errorf("event %d label %q does not match move %d (%q)", i, ev.Label, ev.Index, ev.Input.Moves[ev.Index].Label)
			}
		}
	}
	// Every applied move is exactly one event, and the runner's own
	// counter is the cross-check: an observer that missed a window,
	// or fired twice on one, breaks this.
	if int64(applied) != st.Applied {
		t.Errorf("%d applied events for %d applied moves", applied, st.Applied)
	}
}

// The heuristic tier is Layer A over Layer B, and both halves are
// Tracers, so every window comes back with the layer that answered it
// and — on the Layer B ones — the ranking Decide throws away.
func TestRunnerObserverCarriesTheFunnelTrace(t *testing.T) {
	c, _ := runWithObserver(t, rules.NewFilter(heuristic.New(), nil), aiseat.Config{}, 12)
	layers := map[string]int{}
	for i, ev := range c.events() {
		if !ev.Traced {
			t.Fatalf("event %d is untraced; the heuristic tier implements aiseat.Tracer", i)
		}
		layers[ev.Trace.Layer]++
		if ev.Trace.Layer == "A" && ev.Trace.Rule == "" {
			t.Errorf("event %d says Layer A answered but names no rule", i)
		}
		if ev.Trace.Layer == "A" && ev.Trace.HeuristicIndex != aiseat.Decline {
			t.Errorf("event %d absorbed by Layer A claims heuristic index %d; nobody asked the heuristic", i, ev.Trace.HeuristicIndex)
		}
	}
	if layers["A"] == 0 {
		t.Errorf("no window was absorbed by Layer A over %d events; the absorption rate is supposed to be most of them", c.len())
	}
}

// A policy that is not a Tracer still produces an event; the runner
// fills in what little it knows. RandomPolicy is the one that has a
// name for it.
func TestRunnerObserverLabelsTheRandomPolicy(t *testing.T) {
	c, _ := runWithObserver(t, aiseat.NewRandomPolicy(nil), aiseat.Config{}, 5)
	for i, ev := range c.events() {
		if ev.Traced {
			t.Errorf("event %d claims a policy trace; RandomPolicy has none", i)
		}
		if ev.Trace.Layer != aiseat.TraceLayerRandom {
			t.Errorf("event %d layer %q, want %q", i, ev.Trace.Layer, aiseat.TraceLayerRandom)
		}
	}
}

func TestRunnerObserverClassifiesFallbacks(t *testing.T) {
	boom := errors.New("policy exploded")
	cases := []struct {
		name   string
		policy aiseat.Policy
		cfg    aiseat.Config
		want   string
	}{
		{"error", failing{err: boom}, aiseat.Config{}, aiseat.FallbackPolicyError},
		{"timeout", slow{}, aiseat.Config{MaxThink: 40 * time.Millisecond}, aiseat.FallbackTimeout},
		{"out of range", outOfRange{}, aiseat.Config{}, aiseat.FallbackOutOfRange},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			const want = 3
			c, _ := runWithObserver(t, tc.policy, tc.cfg, want)
			evs := c.events()
			if len(evs) < want {
				t.Fatalf("only %d events", len(evs))
			}
			// Only the first `want` events are guaranteed to predate
			// the cancel that ends the runner; a decision racing that
			// cancel is classified policy-error on ctx.Canceled, which
			// is correct and not what this test is about.
			for i, ev := range evs[:want] {
				if ev.Fallback != tc.want {
					t.Fatalf("event %d fallback %q, want %q (decision %+v, err %v)",
						i, ev.Fallback, tc.want, ev.Decision, ev.DecisionErr)
				}
				if tc.want != aiseat.FallbackOutOfRange && ev.DecisionErr == nil {
					t.Errorf("event %d has fallback %q but no policy error", i, ev.Fallback)
				}
				if ev.Forced != "" {
					t.Errorf("event %d claims the runner forced %q; nothing was rejected", i, ev.Forced)
				}
			}
		})
	}
}

// decliner never wants anything. On a window where it does not hold
// priority that is honoured; where it does, the runner turns it into
// the pass it stood in for and says so.
//
// keepHand is which of those two halves the fixture reaches. A seat
// that declines its mulligan never keeps, so the table never leaves
// the one window that has no pass in it; keeping the hand is what buys
// a window that does.
type decliner struct{ keepHand bool }

func (decliner) Name() string { return "decliner" }

func (d decliner) Decide(_ context.Context, in aiseat.Input) (aiseat.Decision, error) {
	if d.keepHand {
		for i, m := range in.Moves {
			if strings.HasPrefix(m.Label, "Keep hand") {
				return aiseat.Decision{Index: i, Reason: "keep, and want nothing after"}, nil
			}
		}
	}
	return aiseat.Decision{Index: aiseat.Decline, Reason: "nothing here"}, nil
}

// A decline is two different facts depending on whether the seat is
// holding the table up, and the runner has to report which: honoured
// where nobody is waiting, turned into the pass it stood in for where
// somebody is. Both halves, because either one alone is the bug the
// other hides.
//
// This used to be one run asserting `sawPass || sawDecline` over a
// fixture that can only produce declines — the seat declines its
// mulligan, so no window with a pass in it is ever reached — so the
// decline→pass half, which is what the test is named for, was never
// exercised at all (#938).
func TestRunnerObserverRecordsTheDeclineToPass(t *testing.T) {
	t.Run("no pass on offer, so the decline stands", func(t *testing.T) {
		// The mulligan window offers Keep and Mulligan and no pass.
		// Two events: the seat's own first window, and the one it is
		// woken for when the opponent keeps — guaranteed to arrive,
		// because Start subscribed this seat before the opponent
		// existed.
		c, _ := runWithObserver(t, decliner{}, aiseat.Config{}, 2)
		evs := c.events()
		if len(evs) < 2 {
			t.Fatalf("got %d events, want the seat's own window and the one the opponent's keep woke it for", len(evs))
		}
		for i, ev := range evs {
			if ev.Index != aiseat.Decline {
				t.Errorf("event %d dispatched index %d; there is no pass in a mulligan window to turn a decline into", i, ev.Index)
			}
			if ev.Applied {
				t.Errorf("event %d reports a decline as applied", i)
			}
			if ev.Fallback != "" {
				t.Errorf("event %d claims fallback %q; the policy answered exactly what it meant", i, ev.Fallback)
			}
		}
	})

	t.Run("a pass on offer, so the decline becomes it", func(t *testing.T) {
		// Waiting for an APPLIED one, not merely for one: the seat
		// that reaches the dispatcher second in a two-runner race has
		// its pass refused (priority has moved), which is a window
		// that reports the decline→pass perfectly well and never
		// commits anything. That rejection is the ordinary step race
		// between two seats — `isStepRace` in runner_test.go names the
		// same thing — and it is not what this test is about.
		declinePass := func(ev aiseat.DecisionEvent) bool {
			return ev.Fallback == aiseat.FallbackDeclinePass
		}
		c, _ := runWithObserverUntil(t, decliner{keepHand: true}, aiseat.Config{},
			"a decline the runner turned into a pass and played", func(c *collector) bool {
				for _, ev := range c.events() {
					if declinePass(ev) && ev.Applied {
						return true
					}
				}
				return false
			})
		applied := 0
		for i, ev := range c.events() {
			if !declinePass(ev) {
				continue
			}
			if ev.Index < 0 || ev.Index >= len(ev.Input.Moves) {
				t.Fatalf("event %d: a decline that became a pass dispatched index %d of %d moves", i, ev.Index, len(ev.Input.Moves))
			}
			if mv := ev.Input.Moves[ev.Index]; mv.Kind != legal.KindPass {
				t.Errorf("event %d: the decline became %q (%s), which is not a pass", i, mv.Label, mv.Kind)
			}
			if ev.Decision.Index != aiseat.Decline {
				t.Errorf("event %d: the policy's own answer reads %d; the runner overwrote it instead of recording it", i, ev.Decision.Index)
			}
			switch {
			case ev.Applied:
				applied++
			// The two ways a decided window legitimately commits
			// nothing, both of which the event has to say out loud or
			// a decision census cannot balance: the dispatcher refused
			// it, or the runner's context ended before it got there.
			case ev.RejectErr != nil || ev.Forced == aiseat.ForcedCancelled:
			default:
				t.Errorf("event %d: the pass was neither played, refused, nor cancelled: %+v", i, ev)
			}
		}
		// The runner played at least one of them, so the table moved:
		// this is the half that keeps a declining bot from stalling a
		// game.
		if applied == 0 {
			t.Fatalf("no decline→pass reached the table over %d events", c.len())
		}
	})
}

// --- latency ---------------------------------------------------------

func TestStatsLatencyIsOrderedAndCounted(t *testing.T) {
	_, r := runWithObserver(t, heuristic.New(), aiseat.Config{}, 10)
	lat := r.Stats().Latency
	if lat.Count == 0 {
		t.Fatal("no latency samples after ten decisions")
	}
	if !(lat.P50 <= lat.P99 && lat.P99 <= lat.P999 && lat.P999 <= lat.Max) {
		t.Errorf("percentiles out of order: %+v", lat)
	}
	if lat.P50 <= 0 {
		t.Errorf("p50 is %v; a decision takes some time", lat.P50)
	}
}

func TestPercentilesOfIsNearestRankAndDoesNotReorder(t *testing.T) {
	if got := aiseat.PercentilesOf(nil); got != (aiseat.Percentiles{}) {
		t.Errorf("PercentilesOf(nil) = %+v, want the zero value", got)
	}
	in := make([]time.Duration, 0, 100)
	for i := 100; i >= 1; i-- {
		in = append(in, time.Duration(i)*time.Millisecond)
	}
	first := in[0]
	got := aiseat.PercentilesOf(in)
	if in[0] != first {
		t.Errorf("PercentilesOf reordered the caller's slice")
	}
	want := aiseat.Percentiles{
		Count: 100,
		P50:   50 * time.Millisecond,
		P99:   99 * time.Millisecond,
		P999:  100 * time.Millisecond,
		Max:   100 * time.Millisecond,
	}
	if got != want {
		t.Errorf("PercentilesOf = %+v, want %+v", got, want)
	}
	// Every reported value must be a sample that actually happened.
	single := aiseat.PercentilesOf([]time.Duration{7 * time.Second})
	if single.P50 != 7*time.Second || single.Max != 7*time.Second || single.Count != 1 {
		t.Errorf("one sample: %+v", single)
	}
}

// --- the two windows that used to vanish ------------------------------

// A decision paid for and then abandoned because the runner's context
// ended during the pacing hold is still a window the policy decided
// in. Before this it produced latency, a Decide call and no event at
// all, so a decision census silently under-counted every game that
// was stopped rather than finished.
func TestRunnerObserverReportsAWindowCancelledDuringPacing(t *testing.T) {
	room := newRoom(t, 2, 21)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := &collector{}
	// MinThink far longer than the cancel below, so the runner is
	// certain to be inside pace() when the context ends.
	cfg := aiseat.Config{MinThink: 2 * time.Second, Observer: c}
	r := aiseat.Start(ctx, room, room.Game.Seats[0].ID, &scripted{prefer: []string{"Keep hand"}}, cfg, nil, testLogger())
	// The window is decided before it is paced, so a counted decision
	// is the runner on its way into the 2s hold — the state this test
	// needs, said by the runner instead of guessed at with a sleep
	// short enough to lose a race on a busy machine (#848).
	waitFor(t, "the runner to decide its first window", func() bool {
		return r.Stats().Decisions > 0
	})
	cancel()
	<-r.Done()

	evs := c.events()
	if len(evs) != 1 {
		t.Fatalf("got %d events, want exactly the one abandoned window", len(evs))
	}
	ev := evs[0]
	if ev.Forced != aiseat.ForcedCancelled {
		t.Errorf("forced %q, want %q", ev.Forced, aiseat.ForcedCancelled)
	}
	if ev.Applied {
		t.Error("an abandoned window was reported as applied")
	}
	if ev.Seq != 0 {
		t.Errorf("seq %d on a window that committed nothing", ev.Seq)
	}
	if ev.Latency <= 0 || ev.Label == "" {
		t.Errorf("the decision itself was not reported: %+v", ev)
	}
	// The policy's own answer survives alongside the runner's
	// override: the seat decided, it just never got to play.
	if ev.Fallback != "" {
		t.Errorf("fallback %q — the policy answered fine", ev.Fallback)
	}
}

// Forced is a SECOND axis, not a refinement of Fallback. A window
// whose policy returned an out-of-range index and whose fallback was
// then rejected onto the always-legal answer has to report both, or
// the "why did this seat play fail-to-find" question has no answer.
func TestRunnerObserverKeepsTheFallbackWhenItForcesAnAnswer(t *testing.T) {
	room, me := newMyriadRoom(t, 7, 5)
	if pi := aiseat.PassIndex(legal.EnumerateFor(room.Game, me.ID)); pi >= 0 {
		t.Fatal("scenario is wrong: a seat owing a choice must not be offered a pass")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := &collector{}
	pol := &millOnPick{t: t, room: room, seat: me.ID}
	r := aiseat.Start(ctx, room, me.ID, pol, aiseat.Config{
		MaxThink:              time.Second,
		MaxConsecutiveRejects: 2,
		Observer:              c,
	}, nil, testLogger())
	waitFor(t, "the search prompt to clear", func() bool {
		open := true
		room.Game.ReadSnapshot(func() { open = searchChoice(room.Game, me.ID) != nil })
		return !open
	})
	cancel()
	<-r.Done()

	var forced, rejected int
	for _, ev := range c.events() {
		if ev.RejectErr != nil {
			rejected++
			if ev.Applied {
				t.Error("a rejected move was reported as applied")
			}
		}
		if ev.Forced == aiseat.ForcedAlwaysLegal {
			forced++
			if !ev.Applied {
				t.Error("the always-legal answer was forced and then not applied")
			}
			if ev.Label == "" {
				t.Error("the forced answer has no label")
			}
		}
	}
	if rejected < 2 {
		t.Errorf("%d rejections observed; the scenario did not reject anything", rejected)
	}
	if forced != 1 {
		t.Errorf("%d forced always-legal answers observed, want 1", forced)
	}
}

// A Filter over a policy that cannot trace itself must not claim the
// inner policy is a heuristic. It was reporting Layer "B" and an
// index, which over a random policy is a fabricated claim about a
// coin flip.
func TestFilterOverAnUntraceablePolicyNamesIt(t *testing.T) {
	c, _ := runWithObserver(t, rules.NewFilter(aiseat.NewRandomPolicy(nil), nil), aiseat.Config{}, 6)
	sawInner := false
	for i, ev := range c.events() {
		if !ev.Traced {
			t.Fatalf("event %d is untraced; the filter implements aiseat.Tracer", i)
		}
		switch ev.Trace.Layer {
		case "A":
			// Layer A answered; nothing below was asked.
		case aiseat.TraceLayerRandom:
			sawInner = true
		default:
			t.Errorf("event %d layer %q — a filter over the random policy has no Layer B", i, ev.Trace.Layer)
		}
		if ev.Trace.HeuristicIndex != aiseat.Decline {
			t.Errorf("event %d claims heuristic index %d; no heuristic was consulted", i, ev.Trace.HeuristicIndex)
		}
	}
	if !sawInner {
		t.Errorf("no window reached the inner policy over %d events", c.len())
	}
}
