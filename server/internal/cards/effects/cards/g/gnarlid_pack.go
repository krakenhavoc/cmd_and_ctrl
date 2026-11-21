package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gnarlid Pack — Creature — Beast {1}{G}, 2/2 (Worldwake, #5582):
//
//	"Multikicker {1}{G} (You may pay an additional {1}{G} any number of times as
//	 you cast this spell.)
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
		OracleID:                   "be6e835c-822e-45ff-a992-249d4d9781dc",
		Name:                       "Gnarlid Pack",
		Completeness:               effects.CompletenessFull,
		OptionalCosts:              []game.AdditionalCost{effects.Multikicker("{1}{G}", 20)},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 1)},
	})
}
