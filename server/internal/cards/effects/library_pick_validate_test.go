package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_pick_validate_test.go — #998. A library-top pick can now
// carry a rule about the picked SET, not just about each card:
// PutFromLibraryOntoBattlefield.Validate and
// TakeFromLibraryToHand.Validate forward
// game.ChooseCardsPrompt.Validate, so "any number of nonland permanent
// cards with TOTAL MANA VALUE 4 OR LESS from among them" (Ao, the Dawn
// Sky) is expressible over a look at the top seven.
//
// The rule itself is enforced in exactly one place — the engine's
// checkChooseCardsPicksLocked, which the submit path and internal/legal
// both go through — so these tests are about the WIRE: that the field
// reaches the prompt, that a refused answer leaves the prompt open, and
// that the forced-answer shortcut steps aside for a set rule. The bot's
// half is in legal/choose_cards_library_validate_test.go.

// queueLibraryPutWithSetRule applies the primitive with a total-mana-
// value rule over the top `n` cards of the chooser's library.
func queueLibraryPutWithSetRule(t *testing.T, g *game.Game, player uuid.UUID, n, limit int, optional bool, max int) {
	t.Helper()
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: player})
		err := PutFromLibraryOntoBattlefield{
			Player:   player,
			Cards:    g.LookAtTopOfLibraryForEffect(player, n),
			Match:    Nonland(),
			Max:      max,
			Optional: optional,
			Validate: b23TotalManaValueAtMost(limit),
			Label:    "test — total mana value " + string(rune('0'+limit)) + " or less",
			Then:     PutRestOnBottomInRandomOrder,
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	one := plTop(me, "One Drop", "Creature — Bear", "{1}")
	two := plTop(me, "Two Drop", "Creature — Bear", "{2}")
	four := plTop(me, "Four Drop", "Creature — Bear", "{4}")

	queueLibraryPutWithSetRule(t, g, me.ID, 3, 4, true, 0)

	pick := latestChooseCardsFor(g, me.ID)
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
	still := latestChooseCardsFor(g, me.ID)
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
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Error("the prompt should be gone once it is answered")
	}
}

// The empty answer is never refused: Validate is not called for it, so
// "put none of them" stays the answer nothing can take away — which is
// what the enumerator marks AlwaysLegal.
func TestLibraryPutSetRuleNeverRefusesTheEmptyAnswer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_ = plTop(me, "Six Drop", "Creature — Bear", "{6}")
	_ = plTop(me, "Seven Drop", "Creature — Bear", "{7}")

	queueLibraryPutWithSetRule(t, g, me.ID, 2, 4, true, 0)

	pick := latestChooseCardsFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	one := plTop(me, "One Drop", "Creature — Bear", "{1}")
	two := plTop(me, "Two Drop", "Creature — Bear", "{2}")

	// Mandatory, no ceiling: floor == ceiling == both candidates.
	queueLibraryPutWithSetRule(t, g, me.ID, 2, 4, false, 0)

	pick := latestChooseCardsFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	one := plTop(me, "One Drop", "Instant", "{1}")
	two := plTop(me, "Two Drop", "Instant", "{2}")
	four := plTop(me, "Four Drop", "Instant", "{4}")
	handBefore := me.Hand.Size()

	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID})
		err := TakeFromLibraryToHand{
			Player:   me.ID,
			Cards:    g.LookAtTopOfLibraryForEffect(me.ID, 3),
			Max:      0,
			Optional: true,
			Validate: b23TotalManaValueAtMost(4),
			Label:    "test — take any number with total mana value 4 or less",
			Then:     TakeRestOnBottomInRandomOrder,
		}.Apply(ctx)
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose_cards prompt was raised")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{two, four}); !errors.Is(err, game.ErrChoiceSetRejected) {
		t.Fatalf("over-budget set: got %v, want ErrChoiceSetRejected", err)
	}
	if me.Hand.Size() != handBefore {
		t.Fatal("a refused set moved cards to the hand")
	}
	if still := latestChooseCardsFor(g, me.ID); still == nil || still.ID != pick.ID {
		t.Fatal("the refused answer took the prompt with it")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{one, two}); err != nil {
		t.Fatalf("legal set: %v", err)
	}
	if me.Hand.Size() != handBefore+2 {
		t.Errorf("hand size %d, want %d", me.Hand.Size(), handBefore+2)
	}
}
