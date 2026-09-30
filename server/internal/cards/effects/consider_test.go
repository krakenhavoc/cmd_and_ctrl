package effects

import "testing"

const considerOracle = "4c9bcba6-87b5-4fb3-97ee-6fe5b739337d"

// TestConsiderSurveilsThenDraws pins the ordering: the surveil prompt
// must be answered before the draw happens, and the card that goes to
// the graveyard is not the card drawn.
func TestConsiderSurveilsThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Bin Me", "Draw Me")
	hand := me.Hand.Size()
	gy := me.Graveyard.Size()

	castCatalogSpell(t, g, "Consider", "Instant", considerOracle, nil)
	passPriorityAroundTable(t, g)

	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no surveil prompt: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) == 0 {
		t.Fatalf("no cards on the surveil prompt: %+v", c)
	}
	if top, ok := g.LookupCardForEffect(c.ScryCards[0]); !ok || top.Name != "Bin Me" {
		t.Fatalf("surveil should look at the top card first: %+v", c.ScryCards)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)

	if me.Graveyard.Size() != gy+2 { // the surveilled card, plus Consider itself once it resolves
		t.Errorf("graveyard after surveil+resolve: %d, want %d", me.Graveyard.Size(), gy+2)
	}
	if me.Hand.Size() != hand+1 { // Consider left the hand, the drawn card arrived
		t.Errorf("hand size after cast+draw: %d, want %d", me.Hand.Size(), hand+1)
	}
	if top := me.Hand.Cards[me.Hand.Size()-1].Name; top != "Draw Me" {
		t.Errorf("drew %q, want the card left after the surveil", top)
	}
}
