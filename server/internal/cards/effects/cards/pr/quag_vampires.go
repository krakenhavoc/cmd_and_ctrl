package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quag Vampires — Creature — Vampire Rogue {B}, 1/1 (Worldwake, #5582):
//
//	"Multikicker {1}{B} (You may pay an additional {1}{B} any number of times as
//	 you cast this spell.)
//	 Swampwalk (This creature can't be blocked as long as defending player
//	 controls a Swamp.)
//	 This creature enters with a +1/+1 counter on it for each time it
//	 was kicked."
//
// Everflowing Chalice's two primitives: Multikicker for the repeatable
// cost and CountersPerKick for the entry, one +1/+1 counter per kick
// (CR 702.33d). The cap of 20 is the engine's, not the card's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "1d28ed9a-cc4e-48d9-a1a6-7504fc6cc56d",
		Name:                       "Quag Vampires",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"swampwalk"},
		OptionalCosts:              []game.AdditionalCost{effects.Multikicker("{1}{B}", 20)},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 1)},
	})
}
