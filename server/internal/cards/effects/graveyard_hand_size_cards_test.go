package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Praetor's Counsel and Wisdom of Ages (#3247): "return … from your
// graveyard to your hand", the spell exiles itself, and the caster has no
// maximum hand size for the rest of the game.

const (
	praetorsCounselOracle = "83d8eda7-b8d2-479a-b1ec-de56833a3182"
	wisdomOfAgesOracle    = "d9722ec3-aa5e-4c69-8ff0-cef1b1839fa4"
)

func TestPraetorsCounselReturnsTheWholeGraveyardExilesItselfAndLiftsTheCap(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	b := pushGraveyardCardTyped(me, "Old Land", "Land")
	c := pushGraveyardCardTyped(me, "Old Bolt", "Instant")
	theirs := pushGraveyardCardTyped(opp, "Their Bear", "Creature — Bear")

	spell := castCatalogSpell(t, g, "Praetor's Counsel", "Sorcery", praetorsCounselOracle, nil)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{a, b, c} {
		if !me.Hand.Contains(id) {
			t.Errorf("graveyard card %s did not return to hand", id)
		}
	}
	if me.Graveyard.Size() != 0 {
		t.Errorf("graveyard = %d cards, want it empty (the spell exiled itself)", me.Graveyard.Size())
	}
	if !g.Exile.Contains(spell) {
		t.Error("Praetor's Counsel did not exile itself")
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's graveyard was touched")
	}
	if got := mhMax(g, me); got != game.NoMaxHandSize {
		t.Errorf("max = %d, want no maximum", got)
	}
	if got := mhMax(g, opp); got != 7 {
		t.Errorf("an opponent's max = %d, want 7", got)
	}
}

func TestWisdomOfAgesReturnsOnlyInstantsAndSorceries(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := pushGraveyardCardTyped(me, "Old Bolt", "Instant")
	rite := pushGraveyardCardTyped(me, "Old Rite", "Sorcery")
	bear := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	land := pushGraveyardCardTyped(me, "Old Land", "Land")

	spell := castCatalogSpell(t, g, "Wisdom of Ages", "Sorcery", wisdomOfAgesOracle, nil)
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(bolt) || !me.Hand.Contains(rite) {
		t.Error("an instant or sorcery card stayed in the graveyard")
	}
	if !me.Graveyard.Contains(bear) || !me.Graveyard.Contains(land) {
		t.Error("a creature or land card left the graveyard")
	}
	if !g.Exile.Contains(spell) {
		t.Error("Wisdom of Ages did not exile itself")
	}
	if got := mhMax(g, me); got != game.NoMaxHandSize {
		t.Errorf("max = %d, want no maximum", got)
	}
}
