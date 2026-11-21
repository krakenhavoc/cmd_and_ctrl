package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sparring Construct — Artifact Creature — Construct {1}, 1/1 (Dominaria, #5530):
//
//	"When this creature dies, put a +1/+1 counter on target creature you control."
//
// Kujar Seedsculptor's counter on a dies trigger: the creature is chosen
// as the trigger goes on the stack (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc948e19-816b-4c13-a36b-73b3e0472dca",
		Name:         "Sparring Construct",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Sparring Construct — put a +1/+1 counter on target creature you control",
				effects.Cmdr2S46PlusOneCounterOnClauseTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
