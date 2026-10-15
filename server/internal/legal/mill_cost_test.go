package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// mill_cost_test.go — #3845: the enumerator's arms for "Mill N cards"
// as a cost, on both owners of the component. The move is offered only
// while the library can pay it (CR 701.17b, CR 118.3), through the
// predicate the engine refuses with, and every move offered is one the
// dispatcher accepts (#544).

// triadLike is The Warring Triad's shape without the catalog:
// "{T}, Mill a card: <nothing>".
func triadLike() game.Card {
	return game.Card{
		Name:     "Triad Stand-in",
		TypeLine: "Legendary Artifact",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:  "{T}, Mill a card: Do nothing.",
			Cost:   game.AbilityCost{Tap: true, Mill: 1},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	}
}

// millikinLike is Millikin's shape without the catalog.
func millikinLike() game.Card {
	return game.Card{
		Name:     "Millikin Stand-in",
		TypeLine: "Artifact",
		ManaAbilities: []game.ManaAbilityShape{{
			TapCost:  true,
			MillCost: 1,
			Produced: "{C}",
			Label:    "{T}, Mill a card: Add {C}",
		}},
	}
}

func TestMillCostIsGatedOnTheLibraryOnAnActivatedAbility(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	src := battlefieldCard(g, active, triadLike())
	advanceTo(t, g, game.StepPrecombatMain)

	g.WithWriteLock(func() { active.Library.Cards = nil })
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), src); len(acts) != 0 {
		t.Fatalf("empty library: want no move, got %v", labels(acts))
	}
	g.WithWriteLock(func() {
		active.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: active.ID, Controller: active.ID})
	})
	moves := legal.EnumerateFor(g, active.ID)
	if acts := activationsOf(moves, src); len(acts) != 1 {
		t.Fatalf("one card: want one move, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, moves)
}

func TestMillCostIsGatedOnTheLibraryOnAManaAbility(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	src := battlefieldCard(g, active, millikinLike())
	advanceTo(t, g, game.StepPrecombatMain)

	g.WithWriteLock(func() { active.Library.Cards = nil })
	if acts := manaMovesOf(legal.EnumerateFor(g, active.ID), src); len(acts) != 0 {
		t.Fatalf("empty library: want no move, got %v", labels(acts))
	}
	g.WithWriteLock(func() {
		active.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: active.ID, Controller: active.ID})
	})
	moves := legal.EnumerateFor(g, active.ID)
	if acts := manaMovesOf(moves, src); len(acts) != 1 {
		t.Fatalf("one card: want one move, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, moves)
}
