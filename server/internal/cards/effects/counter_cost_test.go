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

// A variable AMONG cost is a shape for a printed kind and a floor of one
// (#3276); the shapes nobody prints stay refused.
func TestRegisterAcceptsOnlyThePrintedVariableAmongShape(t *testing.T) {
	among := func(kind string, n int) game.AbilityCost {
		c := RemoveCountersAmong(kind, n, "artifacts you control", Artifact())
		c.RemoveCounters.Variable = true
		return c
	}
	mustPanic(t, "of any kind from among", func() {
		Register(Spec{
			OracleID: "counter-cost-test-variable-among-any",
			Name:     "Variable Among Any Kind",
			Activated: []ActivatedAbility{{
				Label:  "Remove counters from among artifacts you control: do nothing",
				Cost:   among("", 1),
				Effect: func(*game.Game, *game.StackItem) error { return nil },
			}},
		})
	})
	mustPanic(t, "any number", func() {
		Register(Spec{
			OracleID: "counter-cost-test-variable-among-zero",
			Name:     "Variable Among Zero Floor",
			Activated: []ActivatedAbility{{
				Label:  "Remove any number of counters from among artifacts you control: do nothing",
				Cost:   among("+1/+1", 0),
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

// #3987: "Remove two counters from Ghost-Spider" is Among with no From,
// and only for any kinds and a fixed count of at least two. The other
// spellings are a different cost or one no card prints.
func TestRegisterAcceptsOnlyTheSelfAnyKindsAmongShape(t *testing.T) {
	registerForTest(t, Spec{
		OracleID: "counter-cost-test-self-any-kinds",
		Name:     "Self Any Kinds",
		Activated: []ActivatedAbility{{
			Label:  "Remove two counters from this creature: do nothing",
			Cost:   RemoveCountersOfAnyKindsFromThis(2),
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	selfAmong := func(kind string, n int, variable bool) game.AbilityCost {
		return game.AbilityCost{RemoveCounters: &game.CounterRemovalCost{Counter: kind, N: n, Among: true, Variable: variable}}
	}
	for _, tc := range []struct {
		id, want string
		cost     game.AbilityCost
	}{
		{"counter-cost-test-self-among-kind", "RemoveCountersFromThis", selfAmong("+1/+1", 2, false)},
		{"counter-cost-test-self-among-variable", "variable number", selfAmong("", 1, true)},
		{"counter-cost-test-self-among-one", "one counter of any kind", selfAmong("", 1, false)},
	} {
		mustPanic(t, tc.want, func() {
			Register(Spec{
				OracleID: tc.id,
				Name:     tc.id,
				Activated: []ActivatedAbility{{
					Label:  "Remove counters from this: do nothing",
					Cost:   tc.cost,
					Effect: func(*game.Game, *game.StackItem) error { return nil },
				}},
			})
		})
	}
}
