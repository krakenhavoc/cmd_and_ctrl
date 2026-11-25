package game

import (
	"errors"
	"testing"
)

// discard_special_action_test.go — #6224, CR 116.2e: "You may discard
// Circling Vultures any time you could cast an instant." Doing so is a
// special action: no stack, nothing to respond to, the player keeps
// priority.

const vulturesOracle = "test-circling-vultures"

func withVultures(t *testing.T) {
	t.Helper()
	withCatalogSpecialActions(t, func(id string) []SpecialAction {
		if id == vulturesOracle {
			return []SpecialAction{{Kind: SpecialActionDiscard, Label: "Discard this card"}}
		}
		return nil
	})
}

// TestDiscardVulturesWithPriorityOnAnOpponentsTurn — the acceptance test:
// the owner discards it with priority during an opponent's turn, with a
// spell on the stack, and the stack is untouched (CR 116.2e).
func TestDiscardVulturesWithPriorityOnAnOpponentsTurn(t *testing.T) {
	withVultures(t)
	g := newActiveGame(t)
	me := g.Seats[1] // not the active player
	vultures := seedHandCard(me, "Circling Vultures", vulturesOracle, "Creature — Bird", "{B}")
	g.WithWriteLock(func() {
		g.Stack.PushTop(Card{Name: "Something", TypeLine: "Instant", Owner: g.Seats[0].ID, Controller: g.Seats[0].ID})
		g.Turn.PriorityHolder = 1
	})
	stack := g.Stack.Size()

	if err := g.PerformSpecialAction(me.ID, vultures.InstanceID, SpecialActionDiscard, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("discarding Circling Vultures: %v", err)
	}
	if me.Hand.Contains(vultures.InstanceID) || !me.Graveyard.Contains(vultures.InstanceID) {
		t.Fatal("the Vultures did not go from the hand to the graveyard")
	}
	if g.Stack.Size() != stack {
		t.Errorf("stack = %d, want %d: a special action does not use the stack", g.Stack.Size(), stack)
	}
	if g.Turn.PriorityHolder != 1 {
		t.Errorf("priority moved to %d: the player keeps it (CR 116.2)", g.Turn.PriorityHolder)
	}
	var cause DiscardCause
	for _, ev := range g.Events {
		if ev.Kind == EventDiscardCard && ev.CardID == vultures.InstanceID {
			cause = ev.DiscardCause
		}
	}
	if cause != DiscardCauseSpecialAction {
		t.Errorf("discard cause = %q, want %q: neither an effect's nor a cost", cause, DiscardCauseSpecialAction)
	}
}

// TestDiscardVulturesNeedsPriorityAndNoSplitSecond — "any time you could
// cast an instant": not without priority, not under split second (CR
// 702.61a), and only a card that prints it.
func TestDiscardVulturesNeedsPriorityAndNoSplitSecond(t *testing.T) {
	withVultures(t)
	g := newActiveGame(t)
	me := g.Seats[1]
	vultures := seedHandCard(me, "Circling Vultures", vulturesOracle, "Creature — Bird", "{B}")
	other := seedHandCard(me, "Grizzly Bears", "test-bears", "Creature — Bear", "{1}{G}")

	g.WithWriteLock(func() { g.Turn.PriorityHolder = 0 })
	if err := g.PerformSpecialAction(me.ID, vultures.InstanceID, SpecialActionDiscard, SpecialActionParams{Strict: true}); !errors.Is(err, ErrSpecialActionTiming) {
		t.Fatalf("without priority = %v, want ErrSpecialActionTiming", err)
	}
	g.WithWriteLock(func() {
		g.Turn.PriorityHolder = 1
		g.SplitSecondActive = true
	})
	if err := g.PerformSpecialAction(me.ID, vultures.InstanceID, SpecialActionDiscard, SpecialActionParams{Strict: true}); !errors.Is(err, ErrSpecialActionTiming) {
		t.Fatalf("under split second = %v, want ErrSpecialActionTiming", err)
	}
	g.WithWriteLock(func() { g.SplitSecondActive = false })
	if err := g.PerformSpecialAction(me.ID, other.InstanceID, SpecialActionDiscard, SpecialActionParams{Strict: true}); !errors.Is(err, ErrSpecialActionNotOffered) {
		t.Fatalf("a card that does not print it = %v, want ErrSpecialActionNotOffered", err)
	}
	if !me.Hand.Contains(vultures.InstanceID) || !me.Hand.Contains(other.InstanceID) {
		t.Error("a refused special action discarded something")
	}
}
