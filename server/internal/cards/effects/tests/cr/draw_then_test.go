package testscr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

func TestDrawThenDiscardWaitsForADredgeOffer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	imp := effectstest.GyCardWithOracle(p, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	handBefore := p.Hand.Size()

	effectstest.StartLoot(g, p, 2, 2)

	if effectstest.OptionalOffer(g, p) == nil {
		t.Fatal("dredge was not offered on the first draw")
	}
	if effectstest.DiscardPromptFor(g, p) != nil {
		t.Fatal("the discard was queued while the draw was still waiting on the dredge offer")
	}

	// Decline the first draw's offer, then the second's.
	effectstest.AnswerOptional(t, g, p, false)
	if effectstest.DiscardPromptFor(g, p) != nil {
		t.Fatal("the discard was queued before the second draw")
	}
	effectstest.AnswerOptional(t, g, p, false)

	if got := p.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand %d, want %d: both draws resolve before the discard", got, handBefore+2)
	}
	if effectstest.DiscardPromptFor(g, p) == nil {
		t.Fatal("the discard did not open once the draws were done")
	}
	if !effectstest.ZoneHas(p.Graveyard, imp) {
		t.Error("a declined dredge leaves the card in the graveyard")
	}
}

func TestDrawThenDiscardAfterADredgedDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	imp := effectstest.GyCardWithOracle(p, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	handBefore := p.Hand.Size()

	effectstest.StartLoot(g, p, 1, 1)
	if effectstest.DiscardPromptFor(g, p) != nil {
		t.Fatal("the discard was queued before the dredge answer")
	}
	effectstest.AnswerOptional(t, g, p, true)

	if !effectstest.ZoneHas(p.Hand, imp) {
		t.Error("the Imp did not return to the hand")
	}
	if got := p.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d, want %d", got, handBefore+1)
	}
	if effectstest.DiscardPromptFor(g, p) == nil {
		t.Fatal("the discard did not open after the dredge resolved")
	}
}

func TestDrawThenDiscardWithNoReplacementIsUnchanged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	handBefore := p.Hand.Size()

	effectstest.StartLoot(g, p, 2, 2)

	if got := p.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand %d, want %d", got, handBefore+2)
	}
	if pc := effectstest.DiscardPromptFor(g, p); pc == nil {
		t.Fatal("the discard prompt did not open straight after the draws")
	}
}

func TestDrawThenRunsOnAnEmptyLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	trimLibrary(t, p, 0)

	effectstest.StartLoot(g, p, 1, 1)

	if effectstest.DiscardPromptFor(g, p) == nil {
		t.Fatal("the sentence after \"then\" must still happen when the draw found nothing")
	}
}
