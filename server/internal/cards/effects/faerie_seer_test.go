package effects

import (
	"testing"

	"github.com/google/uuid"
)

const faerieSeerOracle = "b2e65e8b-5f08-4cc2-ab1d-00f8903dbea2"

// TestFaerieSeerEntersAndScries2 checks flying and the ETB scry.
func TestFaerieSeerEntersAndScries2(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Bottom Me", "Keep Me")

	id := castCatalogSpell(t, g, "Faerie Seer", "Creature — Faerie Wizard", faerieSeerOracle, nil)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(id) {
		t.Fatal("Faerie Seer should have resolved onto the battlefield")
	}
	if !effectiveAbilitiesContain(t, g, id, "flying") {
		t.Error("Faerie Seer should have flying")
	}

	c := scryChoiceFor(g, me.ID)
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
