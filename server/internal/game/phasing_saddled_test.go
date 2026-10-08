package game

import (
	"testing"

	"github.com/google/uuid"
)

// TestASaddledMountKeepsTheDesignationThroughAPhaseOutAndIn (#2718) —
// CR 702.171b ends "saddled" at end of turn or when the permanent
// leaves the battlefield. Phasing is not leaving (CR 702.26d), so a
// Mount that phases out and back in during the same turn is still
// saddled; the cleanup sweep then ends it as for any other Mount.
func TestASaddledMountKeepsTheDesignationThroughAPhaseOutAndIn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	mount := vanillaOnBattlefield(g, me.ID, "Test Mount", "Creature — Horse Mount", 2, 2)
	g.WithWriteLock(func() { g.SaddleForEffect(mount, nil) })

	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, mount) })
	g.WithWriteLock(func() { g.phaseInLocked([]uuid.UUID{mount}) })

	var saddled bool
	g.WithWriteLock(func() {
		if c := findBattlefieldCard(g, mount); c != nil {
			saddled = c.Saddled
		}
	})
	if !saddled {
		t.Fatal("a Mount that phased out and back in the same turn lost its saddled designation")
	}

	g.WithWriteLock(func() { g.sweepTurnEndLocked() })
	g.WithWriteLock(func() {
		if c := findBattlefieldCard(g, mount); c != nil {
			saddled = c.Saddled
		}
	})
	if saddled {
		t.Error("the cleanup sweep did not end the designation")
	}
}
