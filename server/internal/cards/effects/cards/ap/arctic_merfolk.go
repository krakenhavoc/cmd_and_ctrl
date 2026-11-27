package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arctic Merfolk — Creature — Merfolk {1}{U}, 1/1 (Planeshift, S82 OLD b641, #5615):
//
//	"Kicker—Return a creature you control to its owner's hand. (You may return a
//	 creature you control to its owner's hand in addition to any other costs as
//	 you cast this spell.)
//	 If this creature was kicked, it enters with a +1/+1 counter on it."
//
// The kicker is a return paid as the spell is cast (CR 702.33a, 601.2h); the
// counter is Baloth Gorger's entry clause read off the cast (CR 614.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "c56671e1-28a6-40c3-bad5-b5e617ddb7da",
		Name:                       "Arctic Merfolk",
		Completeness:               effects.CompletenessFull,
		OptionalCosts:              []game.AdditionalCost{effects.KickerReturnToHand("a creature you control", effects.Creature())},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 1)},
	})
}
