package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const scionOfHalasterOracle = "227ddf6c-d21a-48da-84fb-395c7e096914"

func TestScionOfHalasterIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(scionOfHalasterOracle)
	if !ok || spec.Name != "Scion of Halaster" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Scion of Halaster spec = %+v, registered %v", spec, ok)
	}
}

func scionBoard(t *testing.T, commander bool) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	var cmdr uuid.UUID
	if commander {
		cmdr = effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	} else {
		effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	}
	effectstest.B12Push(g, me.ID, "Scion of Halaster", "Legendary Enchantment — Background", scionOfHalasterOracle, 0, 0)
	if commander {
		// Read the commander through the layers so its grant is in force.
		effectstest.LayeredCard(t, g, cmdr)
	}
	return g, me
}

func TestScionOfHalasterFirstDrawLooksAtTwoPutsOneInTheGraveyardThenDraws(t *testing.T) {
	g, me := scionBoard(t, true)
	ids := effectstest.SeedLibrary(me, "Alpha", "Beta", "Gamma", "Delta")
	hand, gy := me.Hand.Size(), me.Graveyard.Size()

	effectstest.B25Draw(g, me.ID, 1)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no look-at-the-top-two prompt")
	}
	if !effectstest.HasID(c.ChooseCards, ids[0]) || !effectstest.HasID(c.ChooseCards, ids[1]) || len(c.ChooseCards) != 2 {
		t.Errorf("offered %v, want exactly the top two", c.ChooseCards)
	}
	if c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Errorf("pick bounds %d..%d, want exactly one", c.ChooseMin, c.ChooseMax)
	}
	if me.Hand.Size() != hand {
		t.Error("a card was drawn before the pick")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, ids[:1]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Graveyard.Contains(ids[0]) || me.Graveyard.Size() != gy+1 {
		t.Error("the chosen card did not go to the graveyard")
	}
	if !me.Hand.Contains(ids[1]) || me.Hand.Size() != hand+1 {
		t.Error("the other card was not the one drawn")
	}

	// Only the first time each turn: the next draw is a plain draw.
	effectstest.B25Draw(g, me.ID, 1)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("the second draw of the turn asked again")
	}
	if !me.Hand.Contains(ids[2]) {
		t.Error("the second draw did not take the next card")
	}
}

func TestScionOfHalasterPickingTheSecondCardLeavesTheFirstToDraw(t *testing.T) {
	g, me := scionBoard(t, true)
	ids := effectstest.SeedLibrary(me, "Alpha", "Beta", "Gamma")
	effectstest.B25Draw(g, me.ID, 1)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, ids[1:2]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Graveyard.Contains(ids[1]) || !me.Hand.Contains(ids[0]) {
		t.Error("expected Beta in the graveyard and Alpha drawn")
	}
}

func TestScionOfHalasterOnlyHelpsCommanderCreaturesYouOwn(t *testing.T) {
	g, me := scionBoard(t, false)
	ids := effectstest.SeedLibrary(me, "Alpha", "Beta")
	effectstest.B25Draw(g, me.ID, 1)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("a table with no commander creature was asked")
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the plain draw did not happen")
	}
}

func TestScionOfHalasterOneCardLibraryPutsItInTheGraveyardAndDrawsNothing(t *testing.T) {
	g, me := scionBoard(t, true)
	me.Library.Cards = nil
	ids := effectstest.SeedLibrary(me, "Only")
	effectstest.B25Draw(g, me.ID, 1)
	if !me.Graveyard.Contains(ids[0]) || me.Hand.Contains(ids[0]) {
		t.Error("the only card should be in the graveyard, not the hand")
	}
}
