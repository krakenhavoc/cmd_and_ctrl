package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spinal Centipede — Creature — Insect {2}{B}, 3/2 (Guilds of Ravnica, #5525):
//
//	"When this creature dies, put a +1/+1 counter on target creature you
//	 control."
//
// Steadfast Sentry's dies trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "43d44a7f-829a-4f4d-9f71-2f4384ddd803",
		Name:         "Spinal Centipede",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Spinal Centipede — put a +1/+1 counter on target creature you control",
				rfCreatureFCountersOnTarget(1)),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
