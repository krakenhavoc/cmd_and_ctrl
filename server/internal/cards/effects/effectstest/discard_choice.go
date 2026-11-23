package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_choice_test.go — the shared vocabulary for "this card made
// somebody discard" (#651), plus the card-level tests for the two
// orders: a loot (draw, THEN discard) and a rummage (discard, THEN
// draw).
//
// Before #651 an effect discard was a counter in Game.DiscardPending —
// the cleanup step's hand-size map — so every test in this package
// asserted `g.DiscardPending[p.ID] == n` and then walked on. That
// assertion is what encoded the bug: nothing waited for the discard,
// and entering cleanup wiped the map. The discard is now a real
// PendingChoice, so the assertion is "a prompt is open for this seat,
// for this many cards", and the table is gated until it is answered.

const MindRotOracleID = "ad44cf74-b717-48fb-9fa2-77512024d76a"

// discardChoiceFor returns the discard prompt owed by a player right
// now, or nil. A discard is the choose-cards pick over the chooser's
// OWN hand; a test with two hand picks open at once tells them apart
// by Reason.
func DiscardChoiceFor(g *game.Game, player uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c == nil || c.Kind != game.PendingChoiceChooseCards {
			continue
		}
		if c.Chooser == player && c.FromPlayer == player {
			return c
		}
	}
	return nil
}

// discardOwed is how many cards a player has been asked to discard by
// an effect, 0 when no prompt is open for them. It replaces the old
// `g.DiscardPending[id]` read everywhere in this package.
func DiscardOwed(g *game.Game, player uuid.UUID) int {
	if c := DiscardChoiceFor(g, player); c != nil {
		return c.ChooseMax
	}
	return 0
}

// answerDiscard answers a player's discard prompt with named cards.
func AnswerDiscard(t *testing.T, g *game.Game, player uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	c := DiscardChoiceFor(g, player)
	if c == nil {
		t.Fatalf("no discard prompt for %s", player)
	}
	if err := g.ResolveChooseCards(c.ID, player, ids); err != nil {
		t.Fatalf("ResolveChooseCards (discard): %v", err)
	}
}

// discardFromHand answers a player's discard prompt off the top of
// their hand — what a test wants whenever WHICH card is pitched is not
// the thing under test.
func DiscardFromHand(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	c := DiscardChoiceFor(g, player)
	if c == nil {
		t.Fatalf("no discard prompt for %s", player)
	}
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Hand.Size() < c.ChooseMax {
		t.Fatalf("hand too small to answer a %d-card discard", c.ChooseMax)
	}
	ids := make([]uuid.UUID, 0, c.ChooseMax)
	for i := 0; i < c.ChooseMax; i++ {
		ids = append(ids, p.Hand.Cards[i].InstanceID)
	}
	if err := g.ResolveChooseCards(c.ID, player, ids); err != nil {
		t.Fatalf("ResolveChooseCards (discard): %v", err)
	}
}
