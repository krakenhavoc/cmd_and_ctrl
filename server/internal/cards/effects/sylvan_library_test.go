package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sylvan_library_test.go — the chain, end to end.
//
// Sylvan Library is the card that named the gap in #74, and what these
// tests are really pinning is not the card: it is that a prompt can
// queue a prompt, that the second one's content depends on the first
// one's answer, and that nothing is ever asked before the player has
// answered what came before it.

const sylvanLibraryOracle = "92eed395-62ca-4293-882b-8565c40daab5"

// advanceToDrawStepOf walks the turn engine to the given seat's draw
// step, where EventBeginDrawStep fires.
func advanceToDrawStepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 300; i++ {
		if g.Turn.Step == game.StepDraw && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward draw step of seat %d: %v", seat, err)
		}
	}
	t.Fatalf("never reached the draw step of seat %d", seat)
}

func chooseCardsChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// sylvanOpening puts a Sylvan Library on seat 1's battlefield, runs the
// table to that seat's draw step, accepts the "draw two additional
// cards" offer and returns the card-set prompt that follows.
func sylvanOpening(t *testing.T, g *game.Game) (*game.Player, *game.PendingChoice) {
	t.Helper()
	owner := g.Seats[1]
	seedLibrary(owner, "Turn", "Extra1", "Extra2", "Spare1", "Spare2", "Spare3")
	pushPermanentForTest(g, owner.ID, "Sylvan Library", sylvanLibraryOracle, "Enchantment")

	advanceToDrawStepOf(t, g, 1)
	// Link 1: "you may draw two additional cards" — the trigger's own
	// CR 603.4 prompt, offered before the trigger is even built.
	offer := latestTriggerPromptFor(g, owner.ID)
	if offer == nil {
		t.Fatal("Sylvan Library offered no draw-step prompt")
	}
	if err := g.ResolveTriggerPrompt(offer.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	passPriorityAroundTable(t, g)

	pick := chooseCardsChoiceFor(g, owner.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt after the extra draws")
	}
	return owner, pick
}

// TestSylvanLibraryChainsThreePrompts is the headline: three links,
// each one queued by the answer to the one before it, and never two
// open at once.
func TestSylvanLibraryChainsThreePrompts(t *testing.T) {
	g := newCatalogGame(t)
	owner, pick := sylvanOpening(t, g)

	// The candidate set is every card drawn this turn — the turn-based
	// draw plus the two extras — not just the two this ability drew.
	if len(pick.ChooseCards) != 3 {
		t.Fatalf("offered %d candidates, want 3 (the turn draw plus two extras)", len(pick.ChooseCards))
	}
	if pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Errorf("bounds are %d..%d, want exactly 2", pick.ChooseMin, pick.ChooseMax)
	}
	if n := len(g.PendingChoices); n != 1 {
		t.Fatalf("%d prompts open at once, want 1 — the chain must not front-load", n)
	}

	chosen := []uuid.UUID{pick.ChooseCards[0], pick.ChooseCards[1]}
	handBefore := owner.Hand.Size()
	lifeBefore := owner.Life
	if err := g.ResolveChooseCards(pick.ID, owner.ID, chosen); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if owner.Hand.Size() != handBefore || owner.Life != lifeBefore {
		t.Fatal("choosing the cards moved something; the choice alone costs nothing")
	}

	// Link 3, card one. It exists only because of the answer above.
	first := confirmChoiceFor(g, owner.ID)
	if first == nil {
		t.Fatal("no pay-or-put-back prompt after the pick")
	}
	if n := len(g.PendingChoices); n != 1 {
		t.Fatalf("%d prompts open at once, want 1 — the two cards are asked about in turn", n)
	}
	// Pay for the first: 4 life, card stays in hand.
	if err := g.ResolveConfirm(first.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveConfirm(pay): %v", err)
	}
	if owner.Life != lifeBefore-4 {
		t.Errorf("life %d -> %d, want -4", lifeBefore, owner.Life)
	}
	if !owner.Hand.Contains(chosen[0]) {
		t.Error("the paid-for card left the hand")
	}

	// Link 3, card two — queued by the FIRST card's answer, priced
	// against a life total the first answer already changed.
	second := confirmChoiceFor(g, owner.ID)
	if second == nil {
		t.Fatal("the first payment did not queue the second card's question")
	}
	libBefore := owner.Library.Size()
	if err := g.ResolveConfirm(second.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveConfirm(put back): %v", err)
	}
	if owner.Life != lifeBefore-4 {
		t.Errorf("declining cost life: %d, want %d", owner.Life, lifeBefore-4)
	}
	if owner.Hand.Contains(chosen[1]) {
		t.Error("the declined card is still in hand")
	}
	if owner.Library.Size() != libBefore+1 {
		t.Errorf("library is %d, want %d", owner.Library.Size(), libBefore+1)
	}
	// On TOP, not the bottom: the card comes straight back next turn,
	// which is what makes the put-back a tempo cost.
	if top, err := owner.Library.Top(); err != nil || top.InstanceID != chosen[1] {
		t.Error("the put-back card did not land on top of the library")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts left open after the chain finished", len(g.PendingChoices))
	}
}

// TestSylvanLibraryDecliningTheDrawEndsTheChain — the first link is a
// "you may", and saying no must leave nothing behind.
func TestSylvanLibraryDecliningTheDrawEndsTheChain(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	seedLibrary(owner, "Turn", "Extra1", "Extra2", "Spare")
	pushPermanentForTest(g, owner.ID, "Sylvan Library", sylvanLibraryOracle, "Enchantment")

	advanceToDrawStepOf(t, g, 1)
	offer := latestTriggerPromptFor(g, owner.ID)
	if offer == nil {
		t.Fatal("Sylvan Library offered no draw-step prompt")
	}
	handBefore := owner.Hand.Size()
	if err := g.ResolveTriggerPrompt(offer.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveTriggerPrompt(no): %v", err)
	}
	passPriorityAroundTable(t, g)

	if owner.Hand.Size() != handBefore {
		t.Errorf("declining drew %d cards", owner.Hand.Size()-handBefore)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("declining left %d prompts open", len(g.PendingChoices))
	}
}

// TestSylvanLibraryCannotPayLifeItDoesNotHave — CR 119.4. A player who
// cannot pay is not asked a question whose only answer is no; the card
// goes back and the chain moves on, rather than stalling on a prompt
// the engine would refuse (#544).
func TestSylvanLibraryCannotPayLifeItDoesNotHave(t *testing.T) {
	g := newCatalogGame(t)
	owner, pick := sylvanOpening(t, g)
	g.WithWriteLock(func() { owner.Life = 3 })

	chosen := []uuid.UUID{pick.ChooseCards[0], pick.ChooseCards[1]}
	libBefore := owner.Library.Size()
	if err := g.ResolveChooseCards(pick.ID, owner.ID, chosen); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts open, want 0 — an unaffordable payment is not a question",
			len(g.PendingChoices))
	}
	if owner.Life != 3 {
		t.Errorf("life is %d, want 3 — nothing was paid", owner.Life)
	}
	if owner.Library.Size() != libBefore+2 {
		t.Errorf("library is %d, want %d — both cards go back", owner.Library.Size(), libBefore+2)
	}
	for _, id := range chosen {
		if owner.Hand.Contains(id) {
			t.Error("a card that could not be paid for stayed in hand")
		}
	}
}
