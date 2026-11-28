package game

import (
	"testing"

	"github.com/google/uuid"
)

// regeneration_rider_test.go — Debt of Loyalty's "You gain control of that
// creature if it regenerates this way" (CR 701.19a): the shield's rider.

func regenGainingControl(t *testing.T, g *Game, id, player uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.RegenerateGainingControlForEffect(id, player, uuid.New(), "Debt of Loyalty — gain control if it regenerates"); err != nil {
			t.Fatalf("RegenerateGainingControlForEffect: %v", err)
		}
	})
}

func controllerOf(t *testing.T, g *Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	var who uuid.UUID
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c := findBattlefieldCard(g, id)
		if c == nil {
			t.Fatalf("card %s is not on the battlefield", id)
		}
		who = c.Controller
	})
	return who
}

// The rider fires when its own shield replaces a destruction, and not before.
func TestRegenerationRiderGivesControlWhenTheShieldIsUsed(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := regenBear(g, opp.ID)
	regenGainingControl(t, g, id, me.ID)
	if got := controllerOf(t, g, id); got != opp.ID {
		t.Fatalf("control changed before the shield was used: %v", got)
	}
	destroy(t, g, id)
	if findBattlefieldCard(g, id) == nil {
		t.Fatal("the shield did not replace the destruction")
	}
	if got := controllerOf(t, g, id); got != me.ID {
		t.Errorf("the creature is controlled by %v, want the spell's controller", got)
	}
	g.WithWriteLock(func() {
		if n := len(g.regenerationRiderIndexesLocked(findBattlefieldCard(g, id))); n != 0 {
			t.Errorf("%d rider(s) left after the shield was used", n)
		}
	})
}

// A creature never destroyed this turn is not taken.
func TestRegenerationRiderDoesNothingIfNothingIsDestroyed(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := regenBear(g, opp.ID)
	regenGainingControl(t, g, id, me.ID)
	g.WithWriteLock(func() { g.sweepTurnEndLocked() })
	if got := controllerOf(t, g, id); got != opp.ID {
		t.Errorf("the creature changed hands with no regeneration: %v", got)
	}
	if shieldsOn(t, g, id) != 0 {
		t.Error("the shield outlived the turn")
	}
	g.WithWriteLock(func() {
		if n := len(g.regenerationRiderIndexesLocked(findBattlefieldCard(g, id))); n != 0 {
			t.Errorf("%d rider(s) outlived the turn", n)
		}
	})
	destroy(t, g, id)
	if findBattlefieldCard(g, id) != nil {
		t.Error("an expired shield replaced a destruction")
	}
}

// With a plain shield as well, the controller spends the plain one first.
func TestRegenerationRiderIsSpentAfterAPlainShield(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := regenBear(g, opp.ID)
	regenGainingControl(t, g, id, me.ID)
	regenerate(t, g, id)
	destroy(t, g, id)
	if got := controllerOf(t, g, id); got != opp.ID {
		t.Fatalf("the first destruction used the rider shield: controller %v", got)
	}
	destroy(t, g, id)
	if findBattlefieldCard(g, id) == nil {
		t.Fatal("the second shield did not replace the destruction")
	}
	if got := controllerOf(t, g, id); got != me.ID {
		t.Errorf("the second regeneration did not give control: %v", got)
	}
}

// A regeneration on a permanent that has left does nothing.
func TestRegenerationRiderNeedsAPermanent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.RegenerateGainingControlForEffect(uuid.New(), me.ID, uuid.New(), "x"); err == nil {
			t.Error("a regeneration with no permanent returned no error")
		}
	})
}
