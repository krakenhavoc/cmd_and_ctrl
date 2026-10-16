package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slice cav-g2 (#3172): tests for the cards whose caveat this slice removed.

const filigreeFamiliarOracle = "b544f690-e4bf-4a5b-984d-9256518fd574"

// Filigree Familiar prints only the enters lifegain and the dies draw. It has
// no mana ability (#3970): the catalog once invented "{2}, Sacrifice: Add one
// mana of any color", which played stronger than printed.
func TestFiligreeFamiliarHasNoManaAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushCatalogPermanent(g, me.ID, "Filigree Familiar", "Artifact Creature — Fox", filigreeFamiliarOracle, false)
	samiGiveColorless(me, 2)

	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Fatal("the Familiar has no mana ability: the activation should be refused")
	}
	if findBattlefieldCardForTest(g, id) == nil {
		t.Error("a refused activation sacrificed the Familiar")
	}
	if got := dcSortedPool(me); len(got) != 2 {
		t.Errorf("pool = %v: the {2} should be untouched", got)
	}
	if spec, ok := Lookup(filigreeFamiliarOracle); !ok || len(spec.ManaAbilities) != 0 {
		t.Errorf("catalog entry should list no mana abilities: %+v", spec.ManaAbilities)
	}
}
