package testsl

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// testPutBudgetRule is the registered rule for a budget of `limit`.
func testPutBudgetRule(t *testing.T, limit int) effects.PutRule {
	t.Helper()
	if limit != 4 {
		t.Fatalf("no registered put rule for a budget of %d", limit)
	}
	return effectstest.TestPutBudget4
}

// queueLibraryPutWithSetRule applies the primitive with a total-mana-
// value rule over the top `n` cards of the chooser's library.
func queueLibraryPutWithSetRule(t *testing.T, g *game.Game, player uuid.UUID, n, limit int, optional bool, max int) {
	t.Helper()
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Controller: player})
		err := effects.PutFromLibraryOntoBattlefield{
			Player:   player,
			Cards:    g.LookAtTopOfLibraryForEffect(player, n),
			Match:    effects.Nonland(),
			Max:      max,
			Optional: optional,
			Rule:     testPutBudgetRule(t, limit),
			Label:    "test — total mana value " + string(rune('0'+limit)) + " or less",
			ThenTail: effects.PutRestOnBottomInRandomOrderTail,
		}.Apply(ctx)
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})
}

// The core contract: an over-budget set is refused, the prompt is still
// open afterwards (so the player can choose again rather than the table
// wedging), and a set inside the budget is accepted and performed.
func TestLibraryPutSetRuleRefusesOverBudgetAndKeepsAsking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	one := effectstest.PlTop(me, "One Drop", "Creature — Bear", "{1}")
	two := effectstest.PlTop(me, "Two Drop", "Creature — Bear", "{2}")
	four := effectstest.PlTop(me, "Four Drop", "Creature — Bear", "{4}")

	queueLibraryPutWithSetRule(t, g, me.ID, 3, 4, true, 0)

	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose_cards prompt was raised")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 3 {
		t.Fatalf("bounds %d..%d, want 0..3 — \"any number\"", pick.ChooseMin, pick.ChooseMax)
	}

	// 2 + 4 = 6. The bounds allow it; the card does not.
	err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{two, four})
	if !errors.Is(err, game.ErrChoiceSetRejected) {
		t.Fatalf("over-budget set: got %v, want ErrChoiceSetRejected", err)
	}
	if g.Battlefield.Contains(two) || g.Battlefield.Contains(four) {
		t.Fatal("a refused set put cards onto the battlefield")
	}
	still := effectstest.LatestChooseCardsFor(g, me.ID)
	if still == nil || still.ID != pick.ID {
		t.Fatal("the refused answer took the prompt with it — the seat is wedged")
	}

	// 1 + 2 = 3, and the four-drop goes to the bottom with the rest.
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{one, two}); err != nil {
		t.Fatalf("legal set: %v", err)
	}
	if !g.Battlefield.Contains(one) || !g.Battlefield.Contains(two) {
		t.Error("the legal set did not enter")
	}
	if g.Battlefield.Contains(four) {
		t.Error("the four-drop was not picked and must not have entered")
	}
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("the prompt should be gone once it is answered")
	}
}

// The empty answer is never refused: Validate is not called for it, so
// "put none of them" stays the answer nothing can take away — which is
// what the enumerator marks AlwaysLegal.
func TestLibraryPutSetRuleNeverRefusesTheEmptyAnswer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_ = effectstest.PlTop(me, "Six Drop", "Creature — Bear", "{6}")
	_ = effectstest.PlTop(me, "Seven Drop", "Creature — Bear", "{7}")

	queueLibraryPutWithSetRule(t, g, me.ID, 2, 4, true, 0)

	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose_cards prompt was raised")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("declining a \"you may\" with a set rule: %v", err)
	}
}

// A set rule switches the "the only legal answer is every candidate"
// shortcut OFF. With the rule, it is the rule and not the count that
// decides which subsets are answers, so the pick goes through the
// prompt — where it is checked — instead of being performed for the
// player.
func TestLibraryPutSetRulePromptsEvenWhenTheCountWouldForceTheAnswer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	one := effectstest.PlTop(me, "One Drop", "Creature — Bear", "{1}")
	two := effectstest.PlTop(me, "Two Drop", "Creature — Bear", "{2}")

	// Mandatory, no ceiling: floor == ceiling == both candidates.
	queueLibraryPutWithSetRule(t, g, me.ID, 2, 4, false, 0)

	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("a forced pick with a set rule must still be asked, so the set is checked")
	}
	if pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("bounds %d..%d, want 2..2", pick.ChooseMin, pick.ChooseMax)
	}
	if g.Battlefield.Contains(one) || g.Battlefield.Contains(two) {
		t.Fatal("the shortcut performed the pick without checking the rule")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{one, two}); err != nil {
		t.Fatalf("1 + 2 = 3 is inside the budget: %v", err)
	}
	if !g.Battlefield.Contains(one) || !g.Battlefield.Contains(two) {
		t.Error("the answered set did not enter")
	}
}

// The same field on the take-to-hand twin, so the two primitives stay
// clause for clause (#952).
func TestLibraryTakeSetRuleRefusesOverBudgetAndKeepsAsking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	one := effectstest.PlTop(me, "One Drop", "Instant", "{1}")
	two := effectstest.PlTop(me, "Two Drop", "Instant", "{2}")
	four := effectstest.PlTop(me, "Four Drop", "Instant", "{4}")
	handBefore := me.Hand.Size()

	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Controller: me.ID})
		err := effects.TakeFromLibraryToHand{
			Player:   me.ID,
			Cards:    g.LookAtTopOfLibraryForEffect(me.ID, 3),
			Max:      0,
			Optional: true,
			Label:    "test — take any number with total mana value 4 or less",
			Key:      effectstest.TestBudgetTake,
		}.Apply(ctx)
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose_cards prompt was raised")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{two, four}); !errors.Is(err, game.ErrChoiceSetRejected) {
		t.Fatalf("over-budget set: got %v, want ErrChoiceSetRejected", err)
	}
	if me.Hand.Size() != handBefore {
		t.Fatal("a refused set moved cards to the hand")
	}
	if still := effectstest.LatestChooseCardsFor(g, me.ID); still == nil || still.ID != pick.ID {
		t.Fatal("the refused answer took the prompt with it")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{one, two}); err != nil {
		t.Fatalf("legal set: %v", err)
	}
	if me.Hand.Size() != handBefore+2 {
		t.Errorf("hand size %d, want %d", me.Hand.Size(), handBefore+2)
	}
}
