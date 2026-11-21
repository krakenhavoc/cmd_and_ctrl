package testsex

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2650: Falcon Abomination's token carries decayed, and the engine
// reads it — the token can't block, and the Falcon itself can.
func TestFalconAbominationMakesADecayedZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	falcon := effectstest.CastCatalogSpell(t, g, "Falcon Abomination", "Creature — Zombie Bird", "21eea63f-72b2-4155-bfad-f9937a8f8614", nil)
	effectstest.PassPriorityAroundTable(t, g)
	zombies := effectstest.BattlefieldIDsNamed(g, "Zombie")
	if len(zombies) != 1 {
		t.Fatalf("%d Zombies, want 1", len(zombies))
	}
	z := zombies[0]
	if effectstest.ControllerOf(t, g, z) != me.ID {
		t.Error("the Zombie is not the Falcon's controller's")
	}
	if !effectstest.HasEffectiveKeyword(t, g, z, game.KeywordDecayed) {
		t.Error("the Zombie token does not have decayed")
	}
	if c, ok := g.LookupCardForEffect(z); !ok || !game.Restricted(&c, game.CantBlock) {
		t.Error("the decayed Zombie can block")
	}
	if !effectstest.HasEffectiveKeyword(t, g, falcon, "flying") {
		t.Error("the Falcon's printed flying did not reach the effective abilities")
	}
	if c, ok := g.LookupCardForEffect(falcon); !ok || game.Restricted(&c, game.CantBlock) {
		t.Error("the Falcon itself can't block")
	}
}
