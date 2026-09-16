package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_cost_test.go — #625, the catalog half of the counter-removal
// cost component: the constructors, the Plus merge, and the boot-time
// guards in Register. The engine rules are in
// game/counter_cost_test.go; the cards that use it are in
// counter_cost_cards_test.go.

// Plus has to carry the counter component through. Every self-form
// card composes it with {T} ("{T}, Remove a gold counter"), and a Plus
// that dropped the field would ship that ability FREE to repeat —
// stronger than printed, the #259 direction, and silent.
func TestPlusKeepsTheCounterRemovalComponent(t *testing.T) {
	cost := Plus(TapCost(), RemoveCountersFromThis("gold", 1), ManaCost("{1}"))
	if !cost.Tap || cost.Mana != "{1}" {
		t.Fatalf("the other components were lost: %+v", cost)
	}
	rc := cost.RemoveCounters
	if rc == nil {
		t.Fatal("Plus dropped the counter-removal component")
	}
	if rc.Counter != "gold" || rc.N != 1 || rc.From != nil {
		t.Errorf("RemoveCounters = %+v, want 1 gold from the source", *rc)
	}

	other := Plus(RemoveCountersFrom(game.CounterLoyalty, 1, "a planeswalker you control", Planeswalker()), TapCost())
	if other.RemoveCounters == nil || other.RemoveCounters.From == nil {
		t.Fatalf("the other-permanent form lost its From clause: %+v", other.RemoveCounters)
	}
	if other.RemoveCounters.From.Label != "a planeswalker you control" {
		t.Errorf("From label = %q", other.RemoveCounters.From.Label)
	}
}

// Register refuses a counter cost that removes nothing: an ability
// whose "remove 0 counters" is always payable is a free ability.
func TestRegisterRejectsACounterCostOfZero(t *testing.T) {
	mustPanic(t, "removes 0 counters", func() {
		Register(Spec{
			OracleID: "counter-cost-test-zero",
			Name:     "Zero Counters",
			Activated: []ActivatedAbility{{
				Label:  "Remove no counters: do nothing",
				Cost:   RemoveCountersFromThis("gold", 0),
				Effect: func(*game.Game, *game.StackItem) error { return nil },
			}},
		})
	})
	mustPanic(t, "removes -1 counters", func() {
		Register(Spec{
			OracleID: "counter-cost-test-negative",
			Name:     "Negative Counters",
			Activated: []ActivatedAbility{{
				Label:  "Remove minus one counters: do nothing",
				Cost:   RemoveCountersFromThis("gold", -1),
				Effect: func(*game.Game, *game.StackItem) error { return nil },
			}},
		})
	})
}

// An any-kind cost of more than one counter might be paid with two
// DIFFERENT kinds, and one kind choice at announce cannot say so.
// Register refuses it rather than let a card file claim the shape.
func TestRegisterRejectsAnAnyKindCostOfMoreThanOne(t *testing.T) {
	mustPanic(t, "of any kind", func() {
		Register(Spec{
			OracleID: "counter-cost-test-any-two",
			Name:     "Two Of Any Kind",
			Activated: []ActivatedAbility{{
				Label:  "Remove two counters from a creature you control: do nothing",
				Cost:   RemoveCountersFrom("", 2, "a creature you control", Creature()),
				Effect: func(*game.Game, *game.StackItem) error { return nil },
			}},
		})
	})
}
