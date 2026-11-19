package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stronghold Confessor — {B} Creature — Human Cleric 1/1 (Foundations
// Jumpstart):
//
//	"Kicker {3}
//	 Menace
//	 If this creature was kicked, it enters with two +1/+1 counters on it."
//
// Gnarlid Colony's kicker-fed entry counters: two for the one kick.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "e7c200a9-4793-44c8-ba8d-6298dc83bbac",
		Name:                       "Stronghold Confessor",
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"menace"},
		OptionalCosts:              []game.AdditionalCost{Kicker("{3}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{CountersPerKick(game.CounterPlusOne, 2)},
	})
}
