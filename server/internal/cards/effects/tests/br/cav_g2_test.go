package testsbr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Filigree Familiar prints only the enters lifegain and the dies draw. It has
// no mana ability (#3970): the catalog once invented "{2}, Sacrifice: Add one
// mana of any color", which played stronger than printed.
func TestFiligreeFamiliarHasNoManaAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Filigree Familiar", "Artifact Creature — Fox", effectstest.FiligreeFamiliarOracle, false)
	effectstest.SamiGiveColorless(me, 2)

	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Fatal("the Familiar has no mana ability: the activation should be refused")
	}
	if effectstest.FindBattlefieldCardForTest(g, id) == nil {
		t.Error("a refused activation sacrificed the Familiar")
	}
	if got := effectstest.DcSortedPool(me); len(got) != 2 {
		t.Errorf("pool = %v: the {2} should be untouched", got)
	}
	if spec, ok := effects.Lookup(effectstest.FiligreeFamiliarOracle); !ok || len(spec.ManaAbilities) != 0 {
		t.Errorf("catalog entry should list no mana abilities: %+v", spec.ManaAbilities)
	}
}
