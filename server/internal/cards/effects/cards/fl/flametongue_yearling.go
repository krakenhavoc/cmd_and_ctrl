package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flametongue Yearling — Creature — Kavu {R}{R}, 2/1 (Modern Horizons 2,
// #5501):
//
//	"Multikicker {2}
//	 This creature enters with a +1/+1 counter on it for each time it was
//	 kicked.
//	 When this creature enters, it deals damage equal to its power to target
//	 creature."
//
// The counters are entry counters read off the resolving spell (CR 702.33d,
// 614.1c), so the damage — its power as it enters or, if it has left, as it
// last existed (CR 608.2h) — includes them. A Yearling that was not cast has
// no kicks and no counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8181f131-626b-4f9a-812c-9fddaae58025",
		Name:         "Flametongue Yearling",
		Completeness: effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{
			effects.Multikicker("{2}", 20),
		},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 1)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Flametongue Yearling — it deals damage equal to its power to target creature",
					effects.SourceDealsDamageEqualToItsPowerToTarget),
				effects.TargetCreature("target creature")),
		},
	})
}
