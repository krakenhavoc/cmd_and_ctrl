package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kujar Seedsculptor — Creature — Elf Druid {1}{G}, 1/2:
//
//	"When this creature enters, put a +1/+1 counter on target creature you
//	 control."
//
// Bat Colony's counter on the clause target (cmdr2S46PlusOneCounterOnClauseTarget). The Seedsculptor may
// target itself, since it is already on the battlefield when this goes on
// the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93f817c7-2a3d-4346-94c1-3ba8bafb485d",
		Name:         "Kujar Seedsculptor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Kujar Seedsculptor — put a +1/+1 counter on target creature you control",
				effects.Cmdr2S46PlusOneCounterOnClauseTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
