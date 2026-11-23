package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spinal Centipede — Creature — Insect {2}{B}, 3/2 (Guilds of Ravnica, #5525):
//
//	"When this creature dies, put a +1/+1 counter on target creature you
//	 control."
//
// Steadfast Sentry's dies trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43d44a7f-829a-4f4d-9f71-2f4384ddd803",
		Name:         "Spinal Centipede",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Spinal Centipede — put a +1/+1 counter on target creature you control",
				effects.RfCreatureFCountersOnTarget(1)),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
