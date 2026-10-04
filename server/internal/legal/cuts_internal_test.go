package legal

import (
	"testing"

	"github.com/google/uuid"
)

// cuts_internal_test.go — ADR 0122 §6.2, the pieces of the cut report
// no board reaches cheaply: the counting arithmetic, the filtered
// subset walk's two ways of stopping, and the expanded request's hard
// ceiling.

func TestBinomialAndSubsetCount(t *testing.T) {
	for _, tc := range []struct{ n, k, want int }{
		{9, 2, 36}, {24, 1, 24}, {7, 2, 21}, {6, 4, 15}, {5, 0, 1}, {3, 4, 0}, {52, 5, 2598960},
	} {
		if got := binomial(tc.n, tc.k); got != tc.want {
			t.Errorf("C(%d,%d) = %d, want %d", tc.n, tc.k, got, tc.want)
		}
	}
	if got := subsetCount(4, 0, 4); got != 16 {
		t.Errorf("subsets of 4 = %d, want 16", got)
	}
	if got := binomial(400, 200); got != countCap {
		t.Errorf("C(400,200) should saturate at countCap, got %d", got)
	}
}

func uuids(n int) []uuid.UUID {
	out := make([]uuid.UUID, n)
	for i := range out {
		out[i] = uuid.New()
	}
	return out
}

// The walk returns exactly what it returned before the report existed;
// the report only learns what lies past the limit.
func TestAllowedSubsetsWalkLeavesTheAnswerAlone(t *testing.T) {
	pool := uuids(10)
	allowEven := func(set []uuid.UUID) bool { return set[0] == pool[0] || set[0] == pool[2] }
	plain := allowedSubsets(pool, 2, 4, allowEven)
	w := &subsetWalk{}
	walked := allowedSubsetsWalk(pool, 2, 4, allowEven, w)
	if len(plain) != len(walked) {
		t.Fatalf("walk returned %d, plain %d", len(walked), len(plain))
	}
	for i := range plain {
		for j := range plain[i] {
			if plain[i][j] != walked[i][j] {
				t.Fatalf("walk answer %d differs from plain", i)
			}
		}
	}
	if w.spare != 1 || w.untested {
		t.Errorf("want one spare subset found past the limit, got %+v", w)
	}
}

// A rule that refuses everything runs the scan budget dry: the walk
// says subsets were left untested, and claims no spare it never saw.
func TestAllowedSubsetsWalkReportsAnExhaustedScan(t *testing.T) {
	pool := uuids(30)
	w := &subsetWalk{}
	out := allowedSubsetsWalk(pool, 3, 2, func([]uuid.UUID) bool { return false }, w)
	if len(out) != 0 || w.spare != 0 || !w.untested {
		t.Fatalf("want nothing found and the scan cut, got %d, %+v", len(out), w)
	}

	e := &enumerator{report: true, admitted: true}
	e.filteredCombos(pool, 3, 3, 2, func([]uuid.UUID) bool { return false })
	if len(e.cuts) != 1 || e.cuts[0].Cap != CapSubsetScan || e.cuts[0].Omitted != 0 || !e.cuts[0].AtLeast {
		t.Fatalf("want one subset_scan cut of unknown size, got %+v", e.cuts)
	}
}

// An expanded request stops at ExpandCeiling and says how many it
// refused.
func TestExpandedRequestStopsAtTheCeiling(t *testing.T) {
	src := uuid.New()
	e := &enumerator{opts: Options{Source: src}.withDefaults(), report: true}
	e.enter(scope{source: src})
	for range ExpandCeiling + 3 {
		e.add(Move{Source: src})
	}
	if len(e.out) != ExpandCeiling {
		t.Fatalf("kept %d moves, the ceiling is %d", len(e.out), ExpandCeiling)
	}
	if len(e.cuts) != 1 || e.cuts[0].Cap != CapCeiling || e.cuts[0].Omitted != 3 || e.cuts[0].AtLeast {
		t.Fatalf("want an exact ceiling cut of 3, got %+v", e.cuts)
	}
	if e.opts.MaxExpansionPerSource != ExpandCeiling || e.opts.MaxX != ExpandCeiling {
		t.Errorf("naming a card must lift the caps, got %+v", e.opts)
	}
}

// One card's X is searched once per announcement; every search past
// the cap is the same larger X, so the max_x cut is never summed.
func TestMaxXCutsAreNotSummed(t *testing.T) {
	e := &enumerator{report: true, admitted: true}
	e.noteCut(CapMaxX, 1, true)
	e.noteCut(CapMaxX, 1, true)
	e.noteCut(CapPerSource, 1, true)
	e.noteCut(CapPerSource, 1, true)
	for _, c := range e.cuts {
		switch c.Cap {
		case CapMaxX:
			if c.Omitted != 1 {
				t.Errorf("max_x summed to %d", c.Omitted)
			}
		case CapPerSource:
			if c.Omitted != 2 {
				t.Errorf("per_source should sum distinct candidates, got %d", c.Omitted)
			}
		}
	}
}

// Without a report the enumerator files nothing: the bot pays nothing
// for it.
func TestNoReportFilesNothing(t *testing.T) {
	e := &enumerator{admitted: true}
	e.noteCut(CapPerSource, 5, false)
	e.combos(uuids(20), 1, 1, 3, CapPerSource)
	if len(e.cuts) != 0 {
		t.Fatalf("filed %+v without a report", e.cuts)
	}
}
