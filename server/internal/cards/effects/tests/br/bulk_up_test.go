package testsbr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const bulkUpOracle = "fade8af0-5fdb-4237-9dd6-48bfd1d62767"

func TestBulkUpDoublesCurrentPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 3, 3)

	effectstest.CastCatalogSpell(t, g, "Bulk Up", "Instant", bulkUpOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear},
	})
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	id := effectstest.SeedGraveyardCard(t, g, "Bulk Up", "Instant", bulkUpOracle)

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

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
