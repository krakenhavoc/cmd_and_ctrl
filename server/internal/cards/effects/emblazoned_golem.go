package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emblazoned Golem — Artifact Creature — Golem {2}, 1/2 (#2575):
//
//	"Kicker {X} (You may pay an additional {X} as you cast this spell.)
//	 Spend only colored mana on X. No more than one mana of each color
//	 may be spent this way.
//	 If this creature was kicked, it enters with X +1/+1 counters on it."
//
// The kicker's {X} is the cast's one announced X (CR 107.3a), so the
// spend clause restricts the mana paid for it: coloured, one of each
// colour (SpellSpendOnlyColoredOnXOneOfEach), which caps a payable X at
// five. The engine pays it as {W}{U}… for a subset of colours the caster
// can pay, and chooses that subset for them as it chooses which mana
// pays every other symbol (ADR 0040's amendment of 2026-10-22). The
// counters are the CR 614.1c entry clause, reading the spell's X (CR
// 107.3m); an unkicked Golem announces no X and enters with none. A
// kicked X = 0 is legal (the card prints no floor) and adds nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "fbeafc81-1ddb-4b54-8f86-e6afd5f81db8",
		Name:                       "Emblazoned Golem",
		Completeness:               CompletenessFull,
		OptionalCosts:              []game.AdditionalCost{KickerX("{X}", 0)},
		XMatters:                   true,
		SpendOnly:                  SpellSpendOnlyColoredOnXOneOfEach(),
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCountersIfKicked(game.CounterPlusOne)},
	})
}
