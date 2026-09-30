package effects

import (
	"testing"
)

const padeemOracle = "0c7ba712-6a99-4d2f-9242-a2163a11f69c"

// TestPadeemGrantsArtifactsHexproof — every artifact Padeem's
// controller controls carries the hexproof keyword; a creature does
// not.
func TestPadeemGrantsArtifactsHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Padeem, Consul of Innovation", "Legendary Creature — Vedalken Artificer", padeemOracle, false)
	art := pushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", "", false)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)

	if !hasEffectiveKeyword(t, g, art, "hexproof") {
		t.Error("an artifact Padeem's controller controls should have hexproof")
	}
	if hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("a non-artifact creature should not gain hexproof from Padeem")
	}
}

// TestPadeemDrawsOnlyWhileTiedForGreatestArtifactManaValue is the
// intervening-if: Padeem's controller draws when they control the
// biggest (or tied-biggest) artifact on the board, and does not when
// an opponent's is strictly bigger.
func TestPadeemDrawsOnlyWhileTiedForGreatestArtifactManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Padeem, Consul of Innovation", "Legendary Creature — Vedalken Artificer", padeemOracle, false)
	mine := pushCatalogPermanent(g, me.ID, "Mine", "Artifact", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == mine {
				g.Battlefield.Cards[i].ManaCost = "{3}"
			}
		}
	})

	handBefore := me.Hand.Size()
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("tied for greatest artifact MV: hand %d -> %d, want +1", handBefore, me.Hand.Size())
	}

	// Give the opponent a strictly bigger artifact; Padeem's
	// controller no longer ties for the greatest.
	theirs := pushCatalogPermanent(g, opp.ID, "Theirs", "Artifact", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == theirs {
				g.Battlefield.Cards[i].ManaCost = "{9}"
			}
		}
	})
	handBefore = me.Hand.Size()
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore {
		t.Errorf("outsized by an opponent's artifact: hand %d -> %d, want no draw", handBefore, me.Hand.Size())
	}
}
