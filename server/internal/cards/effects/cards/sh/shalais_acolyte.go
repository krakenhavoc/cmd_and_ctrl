package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shalai's Acolyte — Creature — Angel {4}{W}, 3/4 (Dominaria United,
// #5487):
//
//	"Kicker {1}{G} (You may pay an additional {1}{G} as you cast this
//	 spell.)
//	 Flying
//	 If this creature was kicked, it enters with two +1/+1 counters on
//	 it."
//
// Vodalian Serpent's CR 614.1c entry clause read off the cast
// (CountersPerKick); a Shalai's Acolyte put onto the battlefield any other
// way was not kicked and enters with none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "53ec394a-5aaa-4d83-82ec-f00e99902c0d",
		Name:                       "Shalai's Acolyte",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"flying"},
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{1}{G}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 2)},
	})
}
