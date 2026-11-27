package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kavu Aggressor — Creature — Kavu {2}{R}, 3/2 (Invasion, S82 OLD #5616):
//
//	"Kicker {4} (You may pay an additional {4} as you cast this spell.)
//	 This creature can't block.
//	 If this creature was kicked, it enters with a +1/+1 counter on it."
//
// Skyclave Shade's kicker, entry counter and "can't block" (CR 614.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "84c00644-d0e2-418b-b931-4e972ecb3366",
		Name:                       "Kavu Aggressor",
		Completeness:               effects.CompletenessFull,
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{4}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 1)},
		Static:                     []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
