package game

import (
	"reflect"
	"strings"
	"testing"
)

// closure_fields_probe_test.go tests the closure ratchet's walker on
// types built to probe it (#1558), because the property that matters —
// a NEW ROUTE to an already-listed closure type fails — cannot be
// shown on Game without changing Game.

// probeLeaf holds a closure directly: the ReplacementEffect of the
// probe world.
type probeLeaf struct {
	Fn func()
}

// probeListed reaches probeLeaf by the route that is "already listed".
type probeListed struct {
	Leaves []probeLeaf
}

// probeSpec is the TargetSpec of the probe world: it reached no
// closure until somebody gave it a slice of an already-listed type.
type probeSpec struct {
	Label  string
	ZZNew  []probeLeaf
	Plain  int
	Nested *probeSpec
}

// probeCycleA / probeCycleB are a cycle in which only A holds a
// closure. B reaches it through A, and a memoised walk that scores a
// type on the current path as false used to score B as closure-free.
type probeCycleA struct {
	Fn func()
	B  *probeCycleB
}

type probeCycleB struct {
	A []probeCycleA
}

type probeRoot struct {
	Listed probeListed
	Spec   probeSpec
	Cycle  probeCycleB
	Data   []string
}

func probeRoutes() map[string]string {
	out := map[string]string{}
	for _, f := range closureRoutesFrom(reflect.TypeOf(probeRoot{})) {
		out[f.path] = f.via
	}
	return out
}

// TestClosureRatchetSeesANewRouteToAListedType is the #1558 probe: a
// new field whose type is a struct already on the list is a line of
// its own, so it fails the ratchet until it is classified. Keyed by
// the closure-holding type, it produced no new line at all.
func TestClosureRatchetSeesANewRouteToAListedType(t *testing.T) {
	routes := probeRoutes()
	for _, want := range []string{
		"probeLeaf.Fn",       // the closure itself
		"probeListed.Leaves", // the route that was already there
		"probeSpec.ZZNew",    // the NEW route to the same type
		"probeSpec.Nested",   // a struct reaching it through itself
		"probeRoot.Listed",
		"probeRoot.Spec",
	} {
		if _, ok := routes[want]; !ok {
			t.Errorf("route %s missing from %v", want, routes)
		}
	}
	for _, not := range []string{"probeSpec.Label", "probeSpec.Plain", "probeRoot.Data"} {
		if _, ok := routes[not]; ok {
			t.Errorf("%s reaches no closure but was listed", not)
		}
	}
	if via := routes["probeSpec.ZZNew"]; via != "probeRoot.Spec.ZZNew" {
		t.Errorf("probeSpec.ZZNew via = %q, want the shortest route probeRoot.Spec.ZZNew", via)
	}
}

// TestClosureRatchetFollowsACycle: B reaches a closure only through A,
// which is on the path when B is first asked about — the case the old
// memoised walk got wrong.
func TestClosureRatchetFollowsACycle(t *testing.T) {
	routes := probeRoutes()
	for _, want := range []string{"probeRoot.Cycle", "probeCycleB.A", "probeCycleA.B", "probeCycleA.Fn"} {
		if _, ok := routes[want]; !ok {
			t.Errorf("route %s missing from %v", want, routes)
		}
	}
}

// TestClosureClassCeilingsOnlyFall pins the second half of the
// ratchet: a blocker class may not grow, a class that shrank has to
// have its ceiling lowered, and rebuilt / keyed / test-only are not
// counted.
func TestClosureClassCeilingsOnlyFall(t *testing.T) {
	recorded := map[string]string{
		"A.a": "census:StackEffects",
		"A.b": "census:StackEffects",
		"B.a": "transient",
		"C.a": "rebuilt",
		"C.b": "keyed",
		"C.c": "test-only",
	}
	exact := map[string]int{"census:StackEffects": 2, "transient": 1}
	if got := checkClosureClassCeilings(recorded, exact); len(got) != 0 {
		t.Fatalf("exact ceilings: want no problems, got %v", got)
	}

	grew := map[string]int{"census:StackEffects": 1, "transient": 1}
	if got := checkClosureClassCeilings(recorded, grew); len(got) != 1 || !strings.Contains(got[0], "may only shrink") {
		t.Errorf("a class over its ceiling: got %v", got)
	}

	shrank := map[string]int{"census:StackEffects": 3, "transient": 1}
	if got := checkClosureClassCeilings(recorded, shrank); len(got) != 1 || !strings.Contains(got[0], "lower closureClassCeilings") {
		t.Errorf("a class under its ceiling: got %v", got)
	}

	unlisted := map[string]int{"census:StackEffects": 2}
	if got := checkClosureClassCeilings(recorded, unlisted); len(got) != 1 || !strings.Contains(got[0], "transient") {
		t.Errorf("a class with no ceiling is a ceiling of zero: got %v", got)
	}

	gone := map[string]int{"census:StackEffects": 2, "transient": 1, "census:ScopedStatics": 4}
	if got := checkClosureClassCeilings(recorded, gone); len(got) != 1 || !strings.Contains(got[0], "delete its closureClassCeilings entry") {
		t.Errorf("a retired class keeps no ceiling: got %v", got)
	}
}
