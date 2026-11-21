package testsex

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// TestFaerieSeerEntersAndScries2 checks flying and the ETB scry.
func TestFaerieSeerEntersAndScries2(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Bottom Me", "Keep Me")

	id := effectstest.CastCatalogSpell(t, g, "Faerie Seer", "Creature — Faerie Wizard", effectstest.FaerieSeerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(id) {
		t.Fatal("Faerie Seer should have resolved onto the battlefield")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, id, "flying") {
		t.Error("Faerie Seer should have flying")
	}

	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no scry prompt from the ETB: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("scry 2 should offer two cards: %d", len(c.ScryCards))
	}
	bottom, keep := c.ScryCards[0], c.ScryCards[1]
	if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{bottom}, []uuid.UUID{keep}); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if top, ok := g.LookupCardForEffect(keep); !ok || top.Name != "Keep Me" {
		t.Errorf("kept card should still be findable: %+v", top)
	}
}
