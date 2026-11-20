package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guul Draz Mucklord — Creature — Crocodile {2}{B}, 2/3 (Zendikar Rising,
// #5508):
//
//	"When this creature dies, put a +1/+1 counter on target creature you
//	 control."
//
// Steadfast Sentry's dies trigger: the target is chosen as it goes on the
// stack, from the creatures its controller still controls, and
// re-checked as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad62cd58-fe6f-4844-b3dc-5ba21327fc43",
		Name:         "Guul Draz Mucklord",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Guul Draz Mucklord — put a +1/+1 counter on target creature you control",
				rfCreatureFCountersOnTarget(1)),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
