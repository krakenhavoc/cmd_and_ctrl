package testsp

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

const padeemOracle = "0c7ba712-6a99-4d2f-9242-a2163a11f69c"

// TestPadeemGrantsArtifactsHexproof — every artifact Padeem's
// controller controls carries the hexproof keyword; a creature does
// not.
func TestPadeemGrantsArtifactsHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Padeem, Consul of Innovation", "Legendary Creature — Vedalken Artificer", padeemOracle, false)
	art := effectstest.PushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", "", false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)

	if !effectstest.HasEffectiveKeyword(t, g, art, "hexproof") {
		t.Error("an artifact Padeem's controller controls should have hexproof")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("a non-artifact creature should not gain hexproof from Padeem")
	}
}

// TestPadeemDrawsOnlyWhileTiedForGreatestArtifactManaValue is the
// intervening-if: Padeem's controller draws when they control the
// biggest (or tied-biggest) artifact on the board, and does not when
// an opponent's is strictly bigger.
func TestPadeemDrawsOnlyWhileTiedForGreatestArtifactManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Padeem, Consul of Innovation", "Legendary Creature — Vedalken Artificer", padeemOracle, false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Mine", "Artifact", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == mine {
				g.Battlefield.Cards[i].ManaCost = "{3}"
			}
		}
	})

	handBefore := me.Hand.Size()
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("tied for greatest artifact MV: hand %d -> %d, want +1", handBefore, me.Hand.Size())
	}

	// Give the opponent a strictly bigger artifact; Padeem's
	// controller no longer ties for the greatest.
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Theirs", "Artifact", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == theirs {
				g.Battlefield.Cards[i].ManaCost = "{9}"
			}
		}
	})
	handBefore = me.Hand.Size()
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore {
		t.Errorf("outsized by an opponent's artifact: hand %d -> %d, want no draw", handBefore, me.Hand.Size())
	}
}
