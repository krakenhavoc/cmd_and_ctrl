package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slinn Voda, the Rising Deep — Legendary Creature — Leviathan
// {6}{U}{U}, 8/8 (Commander Legends, #5507):
//
//	"Kicker {1}{U}
//	 When Slinn Voda enters, if it was kicked, return all creatures to
//	 their owners' hands except for Merfolk, Krakens, Leviathans,
//	 Octopuses, and Serpents."
//
// Battlewing Mystic's kicked enters trigger (an intervening "if", CR 603.4)
// carrying Whelming Wave's exclusion list, with Merfolk added: the
// subtypes are read effective, so a creature turned into a Merfolk stays.
// Slinn Voda is a Leviathan, so it stays too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "753d98d3-d534-42c1-b728-a4f14c7f29ec",
		Name:          "Slinn Voda, the Rising Deep",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Slinn Voda — kicked: return all creatures to their owners' hands except for Merfolk, Krakens, Leviathans, Octopuses, and Serpents",
					effects.Do(effects.BounceAllMatching{
						Match: effects.Except(effects.Creature(), effects.AnySubtype("Merfolk", "Kraken", "Leviathan", "Octopus", "Serpent")),
					})),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}}),
		},
	})
}
