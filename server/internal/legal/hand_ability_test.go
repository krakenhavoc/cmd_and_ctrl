package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// hand_ability_test.go — the enumerator half of #660. The invariant is
// #544's, one zone over: offer exactly the hand activations the engine
// accepts and nothing when the seat cannot pay, and route them through
// the same cost solve the battlefield loop uses (ADR 0062 Decision 6).

func cyclingHandCard(name, cost string) game.Card {
	return game.Card{
		Name:     name,
		TypeLine: "Land",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:   "Cycling " + cost,
			Cost:    game.AbilityCost{Mana: cost, DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Cycling: true,
			Effect:  func(*game.Game, *game.StackItem) error { return nil },
		}},
	}
}

type discardParams struct {
	DiscardIDs []string `json:"discard_ids"`
}

func discardParamsOf(t *testing.T, m legal.Move) discardParams {
	t.Helper()
	var p discardParams
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", m.Params, err)
	}
	return p
}

// A cycling card in hand is an enumerated activation when the seat can
// pay for it and no move at all when it cannot. The move the
// enumerator offers is one the dispatcher accepts.
func TestCyclingFromHandIsEnumeratedOnlyWhenPayable(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	triome := handCard(active, cyclingHandCard("Ketria Triome", "{2}"))
	advanceTo(t, g, game.StepPrecombatMain)

	// No lands, no floating mana: nothing to pay {2} with.
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), triome); len(acts) != 0 {
		t.Fatalf("a seat with no mana: want no cycling move, got %v", labels(acts))
	}

	battlefieldCard(g, active, basic("Forest", "Forest"))
	battlefieldCard(g, active, basic("Island", "Island"))
	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, triome)
	if len(acts) != 1 {
		t.Fatalf("want exactly one cycling move, got %v", labels(acts))
	}
	// Cycling's cost names no card on the wire: "Discard this card"
	// is the source, not a pick.
	if ids := discardParamsOf(t, acts[0]).DiscardIDs; len(ids) != 0 {
		t.Errorf("cycling move sends discard_ids %v, want none", ids)
	}
	dispatchAll(t, g, active.ID, moves)
}

// A battlefield ability is never offered on a card in hand, and a
// cycling ability is never offered on a permanent (CR 113.6).
func TestAbilityZoneFiltersTheEnumeration(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	battlefieldCard(g, active, basic("Forest", "Forest"))
	battlefieldCard(g, active, basic("Island", "Island"))

	// The same card, with the same ability, on the battlefield.
	onBoard := battlefieldCard(g, active, cyclingHandCard("Ketria Triome", "{2}"))
	// An ordinary battlefield ability, in hand.
	inHand := handCard(active, game.Card{
		Name:     "Sol Ring",
		TypeLine: "Artifact",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:  "{2}: do nothing",
			Cost:   game.AbilityCost{Mana: "{2}"},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	if acts := activationsOf(moves, onBoard); len(acts) != 0 {
		t.Errorf("cycling offered on a permanent: %v", labels(acts))
	}
	if acts := activationsOf(moves, inHand); len(acts) != 0 {
		t.Errorf("a battlefield ability offered on a hand card: %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, moves)
}

// The general "Discard a creature card" component: not offered when
// the hand holds nothing that matches, and when it is offered the move
// names a card the engine accepts.
func TestDiscardCostIsSolvedFromTheHand(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	shaman := battlefieldCard(g, active, game.Card{
		Name:     "Fauna Shaman",
		TypeLine: "Creature — Elf Shaman",
		Power:    2, Toughness: 2,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "Discard a creature card: do nothing",
			Cost: game.AbilityCost{DiscardCards: &game.DiscardCost{
				N: 1, Label: "a creature card",
				Match: func(c game.Card) bool { return c.IsCreature() },
			}},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	advanceTo(t, g, game.StepPrecombatMain)

	// An empty hand cannot pay.
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), shaman); len(acts) != 0 {
		t.Fatalf("empty hand: want no move, got %v", labels(acts))
	}
	// A hand with nothing the clause admits still cannot pay.
	handCard(active, game.Card{Name: "Bolt", TypeLine: "Instant"})
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), shaman); len(acts) != 0 {
		t.Fatalf("no creature card: want no move, got %v", labels(acts))
	}

	bear := handCard(active, creature("Bear", "{1}{G}", 2, 2))
	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, shaman)
	if len(acts) != 1 {
		t.Fatalf("want exactly one move, got %v", labels(acts))
	}
	ids := discardParamsOf(t, acts[0]).DiscardIDs
	if len(ids) != 1 || ids[0] != bear.String() {
		t.Errorf("discard_ids = %v, want the one creature card %v", ids, bear)
	}
	dispatchAll(t, g, active.ID, moves)
}
