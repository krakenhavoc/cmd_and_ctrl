package game

import (
	"testing"

	"github.com/google/uuid"
)

// entry_counter_order_test.go — #1010. TWO counter KINDS on one entry.
//
// The settled ReplacementEvent.EntersWithCounters map used to be
// drained with a bare `range`, and Go randomises map iteration. Each
// kind is placed through AddCounterForEffect, which opens its own
// RepEventCounter window, so the order decided which window opened
// first, which order a CR 616 prompt inside them was asked in, and
// which order the EventCounterPlaced events landed in the log — all
// different on every run and on every replay of the same game.
//
// applyEntryCountersLocked drains in one canonical order: counter
// name, ascending. The order is not a CR 616 choice — see its doc
// comment — and two kinds are latent in the catalog today, which is
// why this is pinned at the engine level rather than on a card.

const testTwoKindOracle = "test-enters-with-two-counter-kinds"

// entryCounterOrder is the sequence of counter kinds the log says
// were put on `id`, in the order they were put on.
func entryCounterOrder(g *Game, id uuid.UUID) []string {
	var out []string
	for _, ev := range g.Events {
		if ev.Kind == EventCounterPlaced && ev.Target == id {
			out = append(out, ev.Label)
		}
	}
	return out
}

func sameOrder(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// THE BUG. Run the same entry many times and the order must be the
// same every time — and be the documented one. 60 runs against a
// two-key map: a random drain has a 2^-60 chance of looking stable.
func TestTwoEntryCounterKindsDrainInOneStableOrder(t *testing.T) {
	// "+1/+1" sorts before "charge": '+' is 0x2B, 'c' is 0x63.
	want := []string{CounterPlusOne, "charge"}
	stubEntryCountersFromCast(t, testTwoKindOracle,
		// Declared in the order a card would print them, which is NOT
		// the order they are drained in: the canonical order is the
		// counter name, so that a re-seed from a different code path
		// (an alternative cost's clause ahead of the card's) cannot
		// change it either.
		xEntryCounters("charge"),
		xEntryCounters(CounterPlusOne),
	)
	for run := 0; run < 60; run++ {
		g := newActiveGame(t)
		id := seedXCreature(t, g, testTwoKindOracle)
		castX(t, g, id, 2)

		got := entryCounterOrder(g, id)
		if !sameOrder(got, want) {
			t.Fatalf("run %d drained %v, want %v", run, got, want)
		}
		if n := counters(t, g, id, CounterPlusOne); n != 2 {
			t.Fatalf("run %d: %d +1/+1 counters, want 2", run, n)
		}
		if n := counters(t, g, id, "charge"); n != 2 {
			t.Fatalf("run %d: %d charge counters, want 2", run, n)
		}
	}
}

// The order is a property of the KINDS, not of the seeding order: the
// same two clauses declared the other way round drain the same way.
// That is what makes a replay of the same game reproduce the log even
// when a different code path seeded the map.
func TestEntryCounterOrderIgnoresTheSeedingOrder(t *testing.T) {
	stubEntryCountersFromCast(t, testTwoKindOracle,
		xEntryCounters(CounterPlusOne),
		xEntryCounters("charge"),
	)
	g := newActiveGame(t)
	id := seedXCreature(t, g, testTwoKindOracle)
	castX(t, g, id, 1)

	if got, want := entryCounterOrder(g, id), []string{CounterPlusOne, "charge"}; !sameOrder(got, want) {
		t.Errorf("drained %v, want %v", got, want)
	}
}

// Each kind opens its OWN window, so a doubler applies to both — the
// composition the order was hiding. Doubling Season on an entry with
// two kinds doubles each of them once, and the log still reads in the
// canonical order.
func TestDoublingSeasonDoublesBothEntryCounterKinds(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	stubEntryCountersFromCast(t, testTwoKindOracle,
		xEntryCounters(CounterPlusOne),
		xEntryCounters("charge"),
	)
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		testEntryDoublerOracle: {counterDoubler("Test Doubling Season")},
	})
	pushReplacementSource(g, testEntryDoublerOracle, me.ID)
	id := seedXCreature(t, g, testTwoKindOracle)

	castX(t, g, id, 3)

	if got := counters(t, g, id, CounterPlusOne); got != 6 {
		t.Errorf("+1/+1 under a doubler: %d, want 6", got)
	}
	if got := counters(t, g, id, "charge"); got != 6 {
		t.Errorf("charge under a doubler: %d, want 6", got)
	}
	if got, want := entryCounterOrder(g, id), []string{CounterPlusOne, "charge"}; !sameOrder(got, want) {
		t.Errorf("drained %v, want %v", got, want)
	}
}

// One kind is still one window and one log line: the canonical order
// changes nothing for the fourteen catalogued cards that declare one.
func TestOneEntryCounterKindIsUnchanged(t *testing.T) {
	stubEntryCountersFromCast(t, testEntryCountersOracle, xEntryCounters(CounterPlusOne))
	g := newActiveGame(t)
	id := seedXCreature(t, g, testEntryCountersOracle)
	castX(t, g, id, 4)

	if got, want := entryCounterOrder(g, id), []string{CounterPlusOne}; !sameOrder(got, want) {
		t.Errorf("drained %v, want %v", got, want)
	}
	if got := counters(t, g, id, CounterPlusOne); got != 4 {
		t.Errorf("%d counters, want 4", got)
	}
}
