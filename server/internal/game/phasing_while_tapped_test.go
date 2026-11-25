package game

import (
	"testing"

	"github.com/google/uuid"
)

// phasing_while_tapped_test.go — #6095: The Pandorica's "It can't phase in
// for as long as The Pandorica remains tapped" (CantPhaseInWhileTapped
// ForEffect) and "that permanent phases in" (PhaseInForEffect). The
// catalog half, with the trigger, is in cards/effects.

func phasedOutNow(g *Game, id uuid.UUID) bool {
	var out bool
	g.ReadSnapshot(func() { out = g.IsPhasedOutForEffect(id) })
	return out
}

func TestCantPhaseInWhileTappedHoldsThroughUntapStepsUntilTheSourceUntaps(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	box := pushTappedPermanent(g, me.ID, "Box", "", "Legendary Artifact", true)
	bear := pushScopedTestCreature(g, opp.ID, 2, 2)
	g.WithWriteLock(func() {
		if err := g.PhaseOutForEffect(box, bear); err != nil {
			t.Fatalf("PhaseOutForEffect: %v", err)
		}
		g.CantPhaseInWhileTappedForEffect(box, bear)
	})
	advanceToStepOfSeat(t, g, 1, StepUpkeep)
	if !phasedOutNow(g, bear) {
		t.Fatal("the permanent phased in during its controller's untap step while the source was tapped")
	}
	g.WithWriteLock(func() {
		if got := g.PhaseInForEffect(bear); len(got) != 0 {
			t.Fatal("PhaseInForEffect phased it in while the bar held")
		}
		if err := g.UntapTargetForEffect(box); err != nil {
			t.Fatalf("UntapTargetForEffect: %v", err)
		}
		if got := g.PhaseInForEffect(bear); len(got) != 1 {
			t.Fatalf("PhaseInForEffect phased in %v once the source untapped, want the permanent", got)
		}
	})
	if phasedOutNow(g, bear) {
		t.Fatal("the permanent is still phased out")
	}
}

// The duration never starts for a source that is not tapped (CR 611.2b):
// the permanent phases in on the ordinary clock.
func TestCantPhaseInWhileTappedNeverStartsForAnUntappedSource(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	box := pushTappedPermanent(g, me.ID, "Box", "", "Legendary Artifact", false)
	bear := pushScopedTestCreature(g, opp.ID, 2, 2)
	g.WithWriteLock(func() {
		_ = g.PhaseOutForEffect(box, bear)
		g.CantPhaseInWhileTappedForEffect(box, bear)
	})
	advanceToStepOfSeat(t, g, 1, StepUpkeep)
	if phasedOutNow(g, bear) {
		t.Fatal("the permanent stayed phased out through its controller's untap step")
	}
}

// The bar names the source's object (CR 400.7): once the source has left
// the battlefield the bar is over, and the permanent phases in.
func TestCantPhaseInWhileTappedEndsWhenTheSourceLeaves(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	box := pushTappedPermanent(g, me.ID, "Box", "", "Legendary Artifact", true)
	bear := pushScopedTestCreature(g, opp.ID, 2, 2)
	g.WithWriteLock(func() {
		_ = g.PhaseOutForEffect(box, bear)
		g.CantPhaseInWhileTappedForEffect(box, bear)
	})
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, box); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	g.WithWriteLock(func() {
		if got := g.PhaseInForEffect(bear); len(got) != 1 {
			t.Fatalf("PhaseInForEffect phased in %v after the source left, want the permanent", got)
		}
	})
}
