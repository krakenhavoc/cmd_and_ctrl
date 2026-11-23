package testst

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const thinkTwiceOracle = "fa85c5a2-8e83-4624-a35a-a0bbf17ecbb4"

func TestThinkTwiceDrawsACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Think Twice", "Instant", thinkTwiceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// castCatalogSpell pushes Think Twice into hand (before+1) then
	// casts it to the stack (back to before); the resolving draw
	// should land the hand at before+1 again.
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand size after cast+draw: got %d, want %d", got, before+1)
	}
}

func TestThinkTwiceFlashback(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.SeedGraveyardCard(t, g, "Think Twice", "Instant", thinkTwiceOracle)
	handBefore := me.Hand.Size()

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand size: got %d, want %d", got, handBefore+1)
	}
	if !g.Exile.Contains(id) {
		t.Errorf("the flashed-back instant should be exiled")
	}
}
