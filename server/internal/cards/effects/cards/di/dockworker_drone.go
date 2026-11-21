package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dockworker Drone — Artifact Creature — Robot {1}{W}, 1/1 (Edge of
// Eternities, #4461):
//
//	"This creature enters with a +1/+1 counter on it.
//	 When this creature dies, put its counters on target creature you
//	 control."
//
// Broodguard Elite's hand-off of last-known counters (CR 122.8, 603.10),
// narrowed to a death, with the entry counter of Afiya Grove.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa704b26-729a-4d0f-b92d-56aadb1e20b9",
		Name:         "Dockworker Drone",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 1, "Dockworker Drone: enters with a +1/+1 counter on it"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Dockworker Drone — put its counters on target creature you control",
				effects.PutItsCountersOnTargetCreature),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
