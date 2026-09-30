package effects

import "testing"

const readTheBonesOracle = "5bf4d8d9-a2b2-4dba-ac05-9d4470a89db2"

// TestReadTheBonesScriesThenDrawsAndLosesLife pins the ordering (scry
// settles before the draw happens) and the unconditional life loss.
func TestReadTheBonesScriesThenDrawsAndLosesLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Bottom Me", "Bottom Too", "Draw One", "Draw Two")
	hand := me.Hand.Size()
	life := me.Life

	castCatalogSpell(t, g, "Read the Bones", "Sorcery", readTheBonesOracle, nil)
	passPriorityAroundTable(t, g)

	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no scry prompt: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("scry 2 should offer two cards: %d", len(c.ScryCards))
	}
	// Put both on the bottom.
	if err := g.ResolveScry(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != hand+2 {
		t.Errorf("hand size after cast+draw two: %d, want %d", me.Hand.Size(), hand+2)
	}
	names := map[string]bool{}
	for i := me.Hand.Size() - 2; i < me.Hand.Size(); i++ {
		names[me.Hand.Cards[i].Name] = true
	}
	if !names["Draw One"] || !names["Draw Two"] {
		t.Errorf("drew the wrong cards after bottoming the scry-2 picks: %+v", names)
	}
	if me.Life != life-2 {
		t.Errorf("life after Read the Bones: %d, want %d", me.Life, life-2)
	}
}
