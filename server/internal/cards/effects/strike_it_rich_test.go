package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const strikeItRichOracle = "c34c17c7-3827-49a2-be25-67f44fdfe150"

func TestStrikeItRichCreatesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := b43TokensNamed(g, me.ID, "Treasure")

	castCatalogSpell(t, g, "Strike It Rich", "Sorcery", strikeItRichOracle, nil)
	passPriorityAroundTable(t, g)

	if got := b43TokensNamed(g, me.ID, "Treasure"); got != before+1 {
		t.Fatalf("Treasure tokens %d -> %d, want +1", before, got)
	}
}

func TestStrikeItRichFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := seedGraveyardCard(t, g, "Strike It Rich", "Sorcery", strikeItRichOracle)
	before := b43TokensNamed(g, me.ID, "Treasure")

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := b43TokensNamed(g, me.ID, "Treasure"); got != before+1 {
		t.Fatalf("Treasure tokens %d -> %d, want +1", before, got)
	}
	if !g.Exile.Contains(id) {
		t.Errorf("the flashed-back sorcery should be exiled")
	}
}
