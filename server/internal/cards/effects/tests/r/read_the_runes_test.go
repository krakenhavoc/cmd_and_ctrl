package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestReadTheRunesDrawsXThenAsksOncePerCard walks both branches in one
// resolution: repetition one is paid with a discard, repetition two
// with a sacrifice.
func TestReadTheRunesDrawsXThenAsksOncePerCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	hand := me.Hand.Size()

	effectstest.CastReadTheRunes(t, g, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	// Taken after the spell has resolved: Read the Runes is itself in
	// the graveyard by the time the first question is asked.
	graves := me.Graveyard.Size()

	// "Draw X cards" happened before any question was asked.
	if got := me.Hand.Size(); got != hand+2 {
		t.Fatalf("hand %d → %d, want +2 drawn before the first question", hand, got)
	}
	first := effectstest.LatestOptionPickFor(g, me.ID)
	if first == nil {
		t.Fatalf("no question was asked: %+v", g.PendingChoices)
	}
	if len(first.PickOptions) != 2 {
		t.Fatalf("options = %d, want 2 (discard, sacrifice)", len(first.PickOptions))
	}

	// Repetition one: discard.
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.DiscardFromHand(t, g, me.ID)
	if got := me.Graveyard.Size(); got != graves+1 {
		t.Errorf("graveyard %d → %d after the discard branch", graves, got)
	}

	// Repetition two: sacrifice. The question is only asked once the
	// first one has been paid.
	second := effectstest.LatestOptionPickFor(g, me.ID)
	if second == nil {
		t.Fatalf("the second question was not asked: %+v", g.PendingChoices)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("the sacrifice branch asked nothing: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{token}); err != nil {
		t.Fatalf("ResolveChooseCards (sacrifice): %v", err)
	}

	if g.Battlefield.Contains(token) {
		t.Error("the chosen permanent was not sacrificed")
	}
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Error("a third question was asked for X = 2")
	}
}

// TestReadTheRunesOffersOnlyTheDiscardWithNoPermanents is CR 608.2's
// "as much as possible": an option the player cannot take is never
// offered, and the discard — the printed default — is always first.
func TestReadTheRunesOffersOnlyTheDiscardWithNoPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]

	effectstest.CastReadTheRunes(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.LatestOptionPickFor(g, me.ID)
	if c == nil {
		t.Fatalf("no question was asked: %+v", g.PendingChoices)
	}
	if len(c.PickOptions) != 1 {
		t.Fatalf("options = %d, want 1 — there is nothing to sacrifice", len(c.PickOptions))
	}
	if c.PickOptions[0].Label != "Discard a card" {
		t.Errorf("the only option is %q, want the discard", c.PickOptions[0].Label)
	}
}

// TestReadTheRunesForZeroDoesNothing: X = 0 draws nothing and asks
// nothing, which is a legal way to cast it.
func TestReadTheRunesForZeroDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()

	effectstest.CastReadTheRunes(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d → %d, want unchanged", hand, got)
	}
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Error("X = 0 asked a question")
	}
}
