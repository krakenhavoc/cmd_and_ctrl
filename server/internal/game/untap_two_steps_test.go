package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// TestSkipNextTwoUntapStepsCountsSteps is #2029's core: a marker for two
// steps keeps the permanent tapped through its controller's next two
// untap steps and releases it at the third.
func TestSkipNextTwoUntapStepsCountsSteps(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	id := pushTappedPermanent(g, p.ID, "Frozen", "", "Creature", true)
	g.WithWriteLock(func() {
		if err := g.SkipNextUntapStepsForEffect(id, p.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	passTurnTo(t, g, 0)
	if !findCard(g, id).Tapped {
		t.Fatal("first untap step must be skipped")
	}
	passTurnTo(t, g, 0)
	if !findCard(g, id).Tapped {
		t.Fatal("second untap step must be skipped")
	}
	passTurnTo(t, g, 0)
	if findCard(g, id).Tapped {
		t.Fatal("third untap step must untap it")
	}
	if len(untapSkipsFor(g, id)) != 0 {
		t.Fatal("marker is used up")
	}
}

// TestTwoStepSkipFollowsControllerNotCaster pins the ruling: with the
// marker keyed to "its controller" (nil player), the count survives a
// control change and is spent at whoever's untap step comes.
func TestTwoStepSkipFollowsControllerNotCaster(t *testing.T) {
	g := newActiveGame(t)
	p, q := g.Seats[0], g.Seats[1]
	id := pushTappedPermanent(g, p.ID, "Frozen", "", "Creature", true)
	g.WithWriteLock(func() {
		_ = g.SkipNextUntapStepsForEffect(id, uuid.Nil, 2)
	})
	passTurnTo(t, g, 0)
	if !findCard(g, id).Tapped {
		t.Fatal("skipped at the first controller's step")
	}
	g.WithWriteLock(func() { cardByIDForUntapTest(g, id).Controller = q.ID })
	passTurnTo(t, g, 0)
	if !findCard(g, id).Tapped {
		t.Fatal("seat 0's step is not its new controller's: nothing is spent")
	}
	passTurnTo(t, g, 1)
	if !findCard(g, id).Tapped {
		t.Fatal("the new controller's step is the second skipped one")
	}
	passTurnTo(t, g, 1)
	if findCard(g, id).Tapped {
		t.Fatal("it untaps once both steps were skipped")
	}
}

// TestOverlappingSkipsRaiseNeverStack: a one-step marker raised by a
// two-step effect, and a repeat of the same effect, never stack.
func TestOverlappingSkipsRaiseNeverStack(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	id := pushTappedPermanent(g, p.ID, "Frozen", "", "Creature", true)
	g.WithWriteLock(func() {
		_ = g.SkipNextUntapForEffect(id, p.ID)
		_ = g.SkipNextUntapStepsForEffect(id, p.ID, 2)
		_ = g.SkipNextUntapStepsForEffect(id, p.ID, 2)
		_ = g.SkipNextUntapStepsForEffect(id, p.ID, 1)
		_ = g.SkipNextUntapStepsForEffect(id, p.ID, 0)
	})
	skips := untapSkipsFor(g, id)
	if len(skips) != 1 || skips[0].Extra != 1 {
		t.Fatalf("skips = %+v, want one marker with Extra 1", skips)
	}
}

func TestUntapSkipExtraSurvivesSnapshot(t *testing.T) {
	b, err := json.Marshal(snapshotUntapSkips([]UntapSkip{{Extra: 2}}))
	if err != nil {
		t.Fatal(err)
	}
	var snap []untapSkipSnapshot
	if err := json.Unmarshal(b, &snap); err != nil {
		t.Fatal(err)
	}
	if out := restoreUntapSkips(snap); len(out) != 1 || out[0].Extra != 2 {
		t.Fatalf("restored %+v", out)
	}
}
