package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const momentaryBlinkOracle = "a3ec6b5d-08ec-4ae0-b1db-c4b87a1849c7"

func TestMomentaryBlinkFlickersAsANewObject(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	castCatalogSpell(t, g, "Momentary Blink", "Instant", momentaryBlinkOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear},
	})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Errorf("the original object should be gone — a blink returns a NEW object")
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" && c.Controller == me.ID {
			found = true
		}
	}
	if !found {
		t.Errorf("a fresh copy of the bear should be back on the battlefield")
	}
}

func TestMomentaryBlinkFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	id := seedGraveyardCard(t, g, "Momentary Blink", "Instant", momentaryBlinkOracle)

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Errorf("the original object should be gone")
	}
	if !g.Exile.Contains(id) {
		t.Errorf("the flashed-back instant should be exiled")
	}
}
