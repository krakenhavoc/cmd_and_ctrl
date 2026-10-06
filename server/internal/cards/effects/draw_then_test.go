package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tests for #2391: a clause printed after a draw ("draw two cards, then
// discard two") waits for a draw that paused on a prompt (CR 608.2c).
// Seat 1 draws, as in draw_instead_test.go.

func discardPromptFor(g *game.Game, p *game.Player) *game.PendingChoice {
	return latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, p.ID)
}

func startLoot(g *game.Game, p *game.Player, draw, discard int) {
	g.WithWriteLock(func() {
		if err := drawThenDiscard(g, p.ID, uuid.Nil, draw, discard, ""); err != nil {
			panic(err)
		}
	})
}

func TestDrawThenDiscardWaitsForADredgeOffer(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[1]
	imp := gyCardWithOracle(p, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	handBefore := p.Hand.Size()

	startLoot(g, p, 2, 2)

	if optionalOffer(g, p) == nil {
		t.Fatal("dredge was not offered on the first draw")
	}
	if discardPromptFor(g, p) != nil {
		t.Fatal("the discard was queued while the draw was still waiting on the dredge offer")
	}

	// Decline the first draw's offer, then the second's.
	answerOptional(t, g, p, false)
	if discardPromptFor(g, p) != nil {
		t.Fatal("the discard was queued before the second draw")
	}
	answerOptional(t, g, p, false)

	if got := p.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand %d, want %d: both draws resolve before the discard", got, handBefore+2)
	}
	if discardPromptFor(g, p) == nil {
		t.Fatal("the discard did not open once the draws were done")
	}
	if !zoneHas(p.Graveyard, imp) {
		t.Error("a declined dredge leaves the card in the graveyard")
	}
}

func TestDrawThenDiscardAfterADredgedDraw(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[1]
	imp := gyCardWithOracle(p, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	handBefore := p.Hand.Size()

	startLoot(g, p, 1, 1)
	if discardPromptFor(g, p) != nil {
		t.Fatal("the discard was queued before the dredge answer")
	}
	answerOptional(t, g, p, true)

	if !zoneHas(p.Hand, imp) {
		t.Error("the Imp did not return to the hand")
	}
	if got := p.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d, want %d", got, handBefore+1)
	}
	if discardPromptFor(g, p) == nil {
		t.Fatal("the discard did not open after the dredge resolved")
	}
}

func TestDrawThenDiscardWithNoReplacementIsUnchanged(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[1]
	handBefore := p.Hand.Size()

	startLoot(g, p, 2, 2)

	if got := p.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand %d, want %d", got, handBefore+2)
	}
	if pc := discardPromptFor(g, p); pc == nil {
		t.Fatal("the discard prompt did not open straight after the draws")
	}
}

func TestDrawThenRunsOnAnEmptyLibrary(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[1]
	trimLibrary(t, p, 0)

	startLoot(g, p, 1, 1)

	if discardPromptFor(g, p) == nil {
		t.Fatal("the sentence after \"then\" must still happen when the draw found nothing")
	}
}
