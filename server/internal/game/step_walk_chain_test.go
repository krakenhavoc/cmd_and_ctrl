package game

import (
	"testing"

	"github.com/google/uuid"
)

// step_walk_chain_test.go — #4162 and #4163: the parked walks are a
// chain, a walk parks directly above the mark its last step began at,
// and the step record lives for one resolution.

func chainKeys(w *stepWalk) []string {
	var out []string
	for ; w != nil; w = w.outer {
		out = append(out, w.then+w.key)
	}
	return out
}

func sameKeys(a, b []string) bool {
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

// A step runs a continuation that parks; then the walk the step belongs
// to parks with the mark taken before the step. The continuation is the
// earlier instruction, so it stays on top. A walk parked before the step
// began (the item's own Effect, waiting behind the prompt the
// continuation answered) stays beneath both.
func TestParkedWalkGoesDirectlyAboveItsMark(t *testing.T) {
	g := &Game{}
	g.parkStepWalkLocked(&stepWalk{key: "effect"}, g.StepWalkMarkForEffect())
	mark := g.StepWalkMarkForEffect() // the resumed walk's step begins
	g.parkStepWalkLocked(&stepWalk{then: "continuation"}, g.StepWalkMarkForEffect())
	shared := g.pausedStepWalk // a clone adopting the chain shares it
	g.parkStepWalkLocked(&stepWalk{key: "resumed"}, mark)
	if got, want := chainKeys(g.pausedStepWalk), []string{"continuation", "resumed", "effect"}; !sameKeys(got, want) {
		t.Fatalf("chain %v, want %v", got, want)
	}
	if got, want := chainKeys(shared), []string{"continuation", "effect"}; !sameKeys(got, want) {
		t.Fatalf("the shared chain became %v, want it left as %v: a park must copy, not relink", got, want)
	}
}

// A mark taken on an empty chain puts the walk at the bottom.
func TestParkedWalkWithAnEmptyMarkGoesToTheBottom(t *testing.T) {
	g := &Game{}
	mark := g.StepWalkMarkForEffect()
	g.parkStepWalkLocked(&stepWalk{then: "inner"}, g.StepWalkMarkForEffect())
	g.parkStepWalkLocked(&stepWalk{then: "outer"}, mark)
	if got, want := chainKeys(g.pausedStepWalk), []string{"inner", "outer"}; !sameKeys(got, want) {
		t.Fatalf("chain %v, want %v", got, want)
	}
}

// A record is read back by item and key, replaced by a second write,
// and dropped as the resolution closes.
func TestStepRecordLivesForOneResolution(t *testing.T) {
	g := &Game{State: StateActive}
	item := &StackItem{ID: uuid.New()}
	other := &StackItem{ID: uuid.New()}
	g.beginResolutionLocked()
	g.RecordStepForEffect(item, "k", StepRecord{IDs: []uuid.UUID{item.ID}, Ns: []int{3}})
	g.RecordStepForEffect(item, "k", StepRecord{Ns: []int{4}})
	if r, ok := g.StepRecordForEffect(item, "k"); !ok || r.N() != 4 || len(r.IDs) != 0 {
		t.Fatalf("record %+v (ok %v), want the second write", r, ok)
	}
	if _, ok := g.StepRecordForEffect(other, "k"); ok {
		t.Fatal("another item read this item's record")
	}
	g.endResolutionLocked()
	if g.holdForOpenResolutionLocked() {
		t.Fatal("nothing holds the resolution open")
	}
	if _, ok := g.StepRecordForEffect(item, "k"); ok {
		t.Fatal("the record outlived its resolution")
	}
}
