package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Picard, Steadfast Captain — Legendary Creature — Human Officer
// {4}{W}{W}, 4/4 (Star Trek Commander, #4974):
//
//	"Whenever Picard enters or attacks, put two +1/+1 counters on
//	 another target creature you control."
//
// One ability with two trigger conditions (WhenThisEntersOrAttacks). The
// target is another creature its controller controls, chosen as the
// trigger goes on the stack; with none, the trigger is removed (CR
// 603.3d), and a target that is gone or no longer yours by resolution
// gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cda786aa-1259-45e8-a834-21c7ad0259a6",
		Name:         "Picard, Steadfast Captain",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEntersOrAttacks("Picard, Steadfast Captain — put two +1/+1 counters on another target creature you control",
				effects.TwoPlusOneCountersOnFirstLegalTarget),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
