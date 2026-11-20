package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sparring Construct — Artifact Creature — Construct {1}, 1/1 (Dominaria, #5530):
//
//	"When this creature dies, put a +1/+1 counter on target creature you control."
//
// Kujar Seedsculptor's counter on a dies trigger: the creature is chosen
// as the trigger goes on the stack (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc948e19-816b-4c13-a36b-73b3e0472dca",
		Name:         "Sparring Construct",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Sparring Construct — put a +1/+1 counter on target creature you control",
				cmdr2S46PlusOneCounterOnClauseTarget),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
