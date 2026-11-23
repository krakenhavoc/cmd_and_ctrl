package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kavu Primarch — Creature — Kavu {3}{G}, 3/3 (Dominaria Remastered):
//
//	"Convoke
//	 Kicker {4}
//	 If this creature was kicked, it enters with four +1/+1 counters on it."
//
// Convoke is the engine's tap-to-pay cost (CR 702.51), kicker an optional
// additional cost (CR 702.33), and the counters are Vodalian Serpent's
// CountersPerKick, read from how the spell was cast as it enters. Convoke
// taps and kicker are independent: a kicked Primarch may be convoked, and
// the kicker mana may be paid by convoke too, as generic mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "09537e9f-6c8e-4849-9d35-d4f3950a64d7",
		Name:          "Kavu Primarch",
		Completeness:  effects.CompletenessFull,
		TapCost:       effects.Convoke(),
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{4}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			effects.CountersPerKick(game.CounterPlusOne, 4),
		},
	})
}
