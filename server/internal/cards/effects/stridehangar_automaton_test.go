package effects

import (
	"testing"
)

const stridehangarAutomatonOracle = "9070c98b-fd01-4eeb-a4ec-fc464946c7c0"

// TestStridehangarAutomatonAddsAThopterWhenAnArtifactTokenIsCreated —
// "those tokens plus an additional Thopter" is an ADDITION, not a
// multiplication: one Treasure creation instruction makes the
// Treasure plus one Thopter, not two Treasures.
func TestStridehangarAutomatonAddsAThopterWhenAnArtifactTokenIsCreated(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	pushTokenReplacementCard(g, stridehangarAutomatonOracle, "Stridehangar Automaton", "Artifact Creature — Construct", me)

	makeTokens(t, g, me, TreasureToken(), 1)

	if got := onBattlefieldNamed(g, "Treasure"); got != 1 {
		t.Errorf("Treasures = %d, want 1 (Stridehangar adds, it does not multiply)", got)
	}
	if got := onBattlefieldNamed(g, "Thopter"); got != 1 {
		t.Errorf("Thopters = %d, want 1 (the additional token)", got)
	}
}

// TestStridehangarAutomatonAnthemBuffsThopters — "Thopters you
// control get +1/+1", checked on a plain Thopter token.
func TestStridehangarAutomatonAnthemBuffsThopters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	pushCatalogPermanent(g, me, "Stridehangar Automaton", "Artifact Creature — Construct", stridehangarAutomatonOracle, false)
	thopter := pushBattlefieldCardWithTimestamp(g, TokenCard("1/1 colorless Thopter artifact with flying"))
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == thopter {
				g.Battlefield.Cards[i].Controller = me
			}
		}
	})
	if p, tg := effectivePower(t, g, thopter), effectiveToughness(t, g, thopter); p != 2 || tg != 2 {
		t.Errorf("Thopter P/T with Stridehangar out = %d/%d, want 2/2", p, tg)
	}
}
