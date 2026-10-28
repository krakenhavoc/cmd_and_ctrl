package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// redirect_order_chooser_test.go — #4389: CR 616.1 gives the ordering to the
// affected object's controller, so a damage instance that reaches creatures
// of two seats keeps one redirection order per chooser.

// One damage instance can carry events on creatures of two seats, each in a
// window of the same redirections, and each seat orders its own. The order
// seat 0 gave for its creature is not seat 1's answer.
func TestRedirectOrderIsKeptPerChooserWithinOneInstance(t *testing.T) {
	g := newActiveGame(t)
	s0, s1 := g.Seats[0], g.Seats[1]
	src := pushKeywordCreature(t, g, s0, 3, 3)
	a := pushKeywordCreature(t, g, s0, 1, 20)
	b := pushKeywordCreature(t, g, s1, 1, 20)
	x := pushKeywordCreature(t, g, s0, 1, 20)
	y := pushKeywordCreature(t, g, s0, 1, 20)
	g.WithWriteLock(func() {
		for _, to := range []uuid.UUID{x, y} {
			if !g.RedirectDamageThisTurnForEffect(DamageRedirection{Controller: s0.ID, ProtectSet: everyCreature, To: to, Label: "Blood"}) {
				t.Fatal("no redirection")
			}
		}
	})
	freshLayers(g)
	var inst DamageInstance
	g.WithWriteLock(func() {
		inst = g.nextDamageInstanceLocked()
		g.openDamageInstance = inst
		_ = g.DealDamageToCreatureForEffect(src, a, 2)
		g.openDamageInstance = 0
	})
	c := openReplacementOrder(g)
	if c == nil || c.Chooser != s0.ID {
		t.Fatalf("want an ordering prompt for seat 0, got %+v", c)
	}
	if err := g.ResolveReplacementOrder(c.ID, c.Chooser, slices.Clone(c.ReplacementEffectIDs)); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		g.openDamageInstance = inst
		_ = g.DealDamageToCreatureForEffect(src, b, 2)
		g.openDamageInstance = 0
	})
	c = openReplacementOrder(g)
	if c == nil || c.Chooser != s1.ID {
		t.Fatalf("seat 1's creature took seat 0's order without being asked: %+v", c)
	}
	if err := g.ResolveReplacementOrder(c.ID, c.Chooser, slices.Clone(c.ReplacementEffectIDs)); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if len(g.redirectOrders) != 2 {
			t.Fatalf("redirect orders %+v, want one per chooser", g.redirectOrders)
		}
	})
}
