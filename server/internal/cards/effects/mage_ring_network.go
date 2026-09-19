package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mage-Ring Network — Land (EDHREC rank 3743):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Put a storage counter on this land.
//	 {T}, Remove any number of storage counters from this land: Add
//	 {C} for each storage counter removed this way."
//
// The storage land, and the card that asked for a VARIABLE counter
// cost (#789). Three abilities:
//
//   - {T}: Add {C}, first so the auto-tapper takes it — the planner
//     reads one ability per permanent, in order, and a Network banked
//     for a big turn should not be spent on a Llanowar Elves.
//   - "{1}, {T}: Put a storage counter" is an ACTIVATED ability, not
//     a mana one: it adds no mana, so it uses the stack as printed
//     (CR 605.1a), and an opponent can respond to the banking.
//   - The payout is a mana ability whose cost removes ANY NUMBER of
//     storage counters. The count is announced at activation — the
//     floor is zero, because "any number" really does include none —
//     and reaches the produced-mana computation through the one
//     paid-cost record: ProducedPerCounterRemoved repeats {C} once
//     per counter that came off.
//
// The count is a fact about the ANNOUNCEMENT, which is what makes the
// card work at all: by the time the mana is minted the counters are
// gone, so nothing on the board could say how many there were.
//
// The auto-tapper never plans the payout — how many counters to spend
// is a decision, and the planner makes none (ADR 0020 §20). It plans
// the plain {C} ability and leaves the bank alone, which is what a
// player wants.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "136596a0-b179-40be-b42d-c0b992621c95",
		Name:         "Mage-Ring Network",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost: ManaAbilityCost{
					Tap:            true,
					RemoveCounters: RemoveCountersXFromThis(game.CounterStorage, 0).RemoveCounters,
				},
				ProducedForPaid: ProducedPerCounterRemoved("{C}"),
				Label:           "Remove any number of storage counters: Add {C} for each",
			},
		},
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}: Put a storage counter on this land.",
			Cost:  Plus(ManaCost("{1}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return AddCounter{
					Target: item.SourceCardID,
					Kind:   game.CounterStorage,
					N:      1,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
