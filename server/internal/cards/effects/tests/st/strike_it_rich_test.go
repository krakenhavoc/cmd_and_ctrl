package testsst

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const strikeItRichOracle = "c34c17c7-3827-49a2-be25-67f44fdfe150"

func TestStrikeItRichCreatesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := effectstest.B43TokensNamed(g, me.ID, "Treasure")

	effectstest.CastCatalogSpell(t, g, "Strike It Rich", "Sorcery", strikeItRichOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.B43TokensNamed(g, me.ID, "Treasure"); got != before+1 {
		t.Fatalf("Treasure tokens %d -> %d, want +1", before, got)
	}
}

func TestStrikeItRichFlashback(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.SeedGraveyardCard(t, g, "Strike It Rich", "Sorcery", strikeItRichOracle)
	before := effectstest.B43TokensNamed(g, me.ID, "Treasure")

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.B43TokensNamed(g, me.ID, "Treasure"); got != before+1 {
		t.Fatalf("Treasure tokens %d -> %d, want +1", before, got)
	}
	if !g.Exile.Contains(id) {
		t.Errorf("the flashed-back sorcery should be exiled")
	}
}
