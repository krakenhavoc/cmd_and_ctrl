package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const bulkUpOracle = "fade8af0-5fdb-4237-9dd6-48bfd1d62767"

func TestBulkUpDoublesCurrentPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 3, 3)

	castCatalogSpell(t, g, "Bulk Up", "Instant", bulkUpOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear},
	})
	passPriorityAroundTable(t, g)

	c, ok := g.LookupCardForEffect(bear)
	if !ok {
		t.Fatalf("bear not found")
	}
	if got := c.CurrentPower(); got != 6 {
		t.Errorf("power = %d, want 6 (doubled from 3)", got)
	}
	if got := c.CurrentToughness(); got != 3 {
		t.Errorf("toughness = %d, want unchanged at 3", got)
	}
}

func TestBulkUpFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	id := seedGraveyardCard(t, g, "Bulk Up", "Instant", bulkUpOracle)

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	c, ok := g.LookupCardForEffect(bear)
	if !ok {
		t.Fatalf("bear not found")
	}
	if got := c.CurrentPower(); got != 4 {
		t.Errorf("power = %d, want 4 (doubled from 2)", got)
	}
	if !g.Exile.Contains(id) {
		t.Errorf("the flashed-back instant should be exiled")
	}
}
