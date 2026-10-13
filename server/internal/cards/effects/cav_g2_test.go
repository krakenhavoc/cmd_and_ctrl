package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slice cav-g2 (#3172): tests for the cards whose caveat this slice removed.

const filigreeFamiliarOracle = "b544f690-e4bf-4a5b-984d-9256518fd574"

// Filigree Familiar's "{2}, Sacrifice: Add one mana of any color" is a mana
// ability: it needs the {2}, sacrifices the Fox, adds the chosen colour, and
// the dies trigger still draws afterwards.
func TestFiligreeFamiliarSacrificeForAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushCatalogPermanent(g, me.ID, "Filigree Familiar", "Artifact Creature — Fox", filigreeFamiliarOracle, false)

	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Fatal("no {2} available: the activation should be refused")
	}
	if findBattlefieldCardForTest(g, id) == nil {
		t.Fatal("a refused activation sacrificed the Familiar")
	}
	samiGiveColorless(me, 2)
	handBefore := me.Hand.Size()
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if findBattlefieldCardForTest(g, id) != nil {
		t.Error("the Familiar should be sacrificed")
	}
	if got := dcSortedPool(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want [G]: the {2} paid, one green added", got)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("dies trigger drew %d cards, want 1", got)
	}
}
