package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rest_in_peace_test.go — the proof for #931. Rest in Peace and
// Leyline of the Void are one sentence, "if a card would be put into a
// graveyard from anywhere, exile it instead", and the word that does
// the work is ANYWHERE: the card is only correct if every route into a
// graveyard opens the CR 614 window. The two that did not were a
// library search with a graveyard destination and surveil's graveyard
// leg, which is why these cards sat on #383's skip list until the
// routes were fixed.
//
// So the tests are one per route rather than one per card: search,
// surveil, mill, a discard and a creature dying. Each of them is a
// different mover in the engine and the same sentence on the card.

const (
	restInPeaceOracle      = "087f9ad7-e74f-40e2-8102-1ed2925d0418"
	leylineOfTheVoidOracle = "f4e32fc1-1b8d-441e-8e76-71f19f98e925"
)

// castRestInPeace casts it from the active seat's hand and lets it
// resolve, so its enters-the-battlefield sweep really fires.
func castRestInPeace(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Rest in Peace", "Enchantment", restInPeaceOracle, nil)
	passPriorityAroundTable(t, g)
	return id
}

// millOne mills one card off a player's library and returns its ID,
// or uuid.Nil when the library was empty.
func millOne(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	top, err := p.Library.Top()
	if err != nil {
		t.Fatalf("library is empty: %v", err)
	}
	id := top.InstanceID
	g.WithWriteLock(func() {
		if _, err := g.MillToZoneForEffect(p.ID, 1, game.ZoneGraveyard); err != nil {
			t.Fatalf("MillToZoneForEffect: %v", err)
		}
	})
	return id
}

// TestRestInPeaceExilesAnEntombedCard is the issue's first route: a
// search that puts the card into a graveyard. Before #931 the take was
// a raw move and Entomb walked straight past the enchantment.
func TestRestInPeaceExilesAnEntombedCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Reanimation Target", "Something Else")
	castRestInPeace(t, g)

	castCatalogSpell(t, g, "Entomb", "Instant", b02bEntombOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "Reanimation Target")

	if b02bGraveyardHasNamed(me, "Reanimation Target") {
		t.Error("the tutored card reached the graveyard; Rest in Peace exiles it instead")
	}
	if !exileHasNamed(g, "Reanimation Target") {
		t.Error("the tutored card is in exile")
	}
}

// TestRestInPeaceExilesASurveilledCard is the second route. Surveil
// bins by choice rather than by count, and the card it bins is put
// into a graveyard like any other.
func TestRestInPeaceExilesASurveilledCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castRestInPeace(t, g)
	seedLibrary(me, "Fodder", "Keeper")

	playLandFromHand(t, g, "Undercity Sewers", undercitySewersOracle)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}

	if b02bGraveyardHasNamed(me, "Fodder") {
		t.Error("the surveilled card reached the graveyard; Rest in Peace exiles it instead")
	}
	if !exileHasNamed(g, "Fodder") {
		t.Error("the surveilled card is in exile")
	}
}

// TestRestInPeaceExilesAMilledCard is the route that has gone through
// the window since #529 — it is here because "from anywhere" is a
// claim about all of them, and a regression in one mover is invisible
// from the others.
func TestRestInPeaceExilesAMilledCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castRestInPeace(t, g)

	milled := millOne(t, g, me)
	if me.Graveyard.Contains(milled) {
		t.Error("a milled card reached the graveyard under Rest in Peace")
	}
	if !g.Exile.Contains(milled) {
		t.Error("the milled card is in exile")
	}
}

// TestRestInPeaceExilesADyingCreature is the "from anywhere" half that
// is a judgement rather than a route: the battlefield is a zone like
// any other, so a creature that would die is exiled instead and never
// died at all (CR 700.4).
func TestRestInPeaceExilesADyingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castRestInPeace(t, g)
	bear := seedPermanentWithOracle(g, me.ID, "Bears", "Creature — Bear", "")

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})

	if me.Graveyard.Contains(bear) {
		t.Error("the destroyed creature reached the graveyard; Rest in Peace exiles it instead")
	}
	if !g.Exile.Contains(bear) {
		t.Error("the destroyed creature is in exile")
	}
}

// TestRestInPeaceExilesAllGraveyardsAsItEnters is the other half of
// the card — the one-shot sweep, every seat's graveyard including its
// controller's.
func TestRestInPeaceExilesAllGraveyardsAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := millOne(t, g, me)
	theirs := millOne(t, g, them)
	if !me.Graveyard.Contains(mine) || !them.Graveyard.Contains(theirs) {
		t.Fatal("both graveyards are seeded before the enchantment arrives")
	}

	castRestInPeace(t, g)

	if me.Graveyard.Size() != 0 || them.Graveyard.Size() != 0 {
		t.Error("the entry sweep empties every graveyard, the controller's included")
	}
	if !g.Exile.Contains(mine) || !g.Exile.Contains(theirs) {
		t.Error("the swept cards are in exile")
	}
}

// TestLeylineOfTheVoidExilesOnlyOpponentsCards is the one clause that
// separates the two cards, and the reason the builder takes a flag
// rather than the cards taking a copy each.
func TestLeylineOfTheVoidExilesOnlyOpponentsCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	castCatalogSpell(t, g, "Leyline of the Void", "Enchantment", leylineOfTheVoidOracle, nil)
	passPriorityAroundTable(t, g)

	mine := millOne(t, g, me)
	theirs := millOne(t, g, them)

	if !me.Graveyard.Contains(mine) {
		t.Error("my own graveyard still works: Leyline names an OPPONENT's graveyard")
	}
	if g.Exile.Contains(mine) {
		t.Error("my own milled card was exiled; Leyline is the one-sided version")
	}
	if them.Graveyard.Contains(theirs) {
		t.Error("the opponent's milled card reached their graveyard")
	}
	if !g.Exile.Contains(theirs) {
		t.Error("the opponent's milled card is in exile")
	}
}

// exileHasNamed reports whether a card of that name is in exile.
func exileHasNamed(g *game.Game, name string) bool {
	for _, c := range g.Exile.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}
