package legal_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// library_costs_test.go — ADR 0109 §7 (#1902): the enumerator's arms
// for "Put a card from your hand on top of your library", "Exile the
// top N cards of your library" and owner decision 3's random discard.
// One payment each, gated on the predicates the engine refuses on, and
// every move offered is one the dispatcher accepts (#544).

type topParams struct {
	TopIDs     []string `json:"top_ids"`
	DiscardIDs []string `json:"discard_ids"`
}

func topParamsOf(t *testing.T, m legal.Move) topParams {
	t.Helper()
	var p topParams
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", m.Params, err)
	}
	return p
}

func lcSource(g *game.Game, p *game.Player, cost game.AbilityCost) game.Card {
	return game.Card{
		Name: "Cost Source", TypeLine: "Enchantment",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:  "Cost: do nothing",
			Cost:   cost,
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	}
}

func TestPutOnTopCostIsSolvedFromTheHand(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	src := battlefieldCard(g, active, lcSource(g, active, game.AbilityCost{PutFromHandOnLibraryTop: 1}))
	advanceTo(t, g, game.StepPrecombatMain)

	if acts := activationsOf(legal.EnumerateFor(g, active.ID), src); len(acts) != 0 {
		t.Fatalf("empty hand: want no move, got %v", labels(acts))
	}
	card := handCard(active, game.Card{Name: "Bolt", TypeLine: "Instant"})
	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, src)
	if len(acts) != 1 {
		t.Fatalf("want one move, got %v", labels(acts))
	}
	if ids := topParamsOf(t, acts[0]).TopIDs; len(ids) != 1 || ids[0] != card.String() {
		t.Errorf("top_ids = %v, want [%v]", ids, card)
	}
	dispatchAll(t, g, active.ID, moves)
}

func TestLibraryExileCostIsGatedOnTheLibrary(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	src := battlefieldCard(g, active, lcSource(g, active, game.AbilityCost{ExileFromLibraryTop: 4}))
	advanceTo(t, g, game.StepPrecombatMain)

	g.WithWriteLock(func() { active.Library.Cards = active.Library.Cards[:3] })
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), src); len(acts) != 0 {
		t.Fatalf("three cards: want no move, got %v", labels(acts))
	}
	g.WithWriteLock(func() {
		for i := 0; i < 2; i++ {
			active.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: active.ID, Controller: active.ID})
		}
	})
	moves := legal.EnumerateFor(g, active.ID)
	if acts := activationsOf(moves, src); len(acts) != 1 {
		t.Fatalf("five cards: want one move, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, moves)
}

func TestRandomDiscardCostIsGatedOnTheHand(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	src := battlefieldCard(g, active, lcSource(g, active, game.AbilityCost{
		DiscardCards: &game.DiscardCost{N: 2, Label: "two cards at random", Random: true}}))
	advanceTo(t, g, game.StepPrecombatMain)

	handCard(active, game.Card{Name: "Bolt", TypeLine: "Instant"})
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), src); len(acts) != 0 {
		t.Fatalf("a hand of one: want no move, got %v", labels(acts))
	}
	handCard(active, game.Card{Name: "Shock", TypeLine: "Instant"})
	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, src)
	if len(acts) != 1 {
		t.Fatalf("a hand of two: want one move, got %v", labels(acts))
	}
	if p := topParamsOf(t, acts[0]); len(p.DiscardIDs) != 0 {
		t.Errorf("a random discard named cards: %v", p.DiscardIDs)
	}
	if !strings.Contains(acts[0].Label, "at random") {
		t.Errorf("label %q does not say the discard is random", acts[0].Label)
	}
	dispatchAll(t, g, active.ID, moves)
}
