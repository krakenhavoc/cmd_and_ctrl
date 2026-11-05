package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "cda786aa-1259-45e8-a834-21c7ad0259a6",
		Name:         "Picard, Steadfast Captain",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEntersOrAttacks("Picard, Steadfast Captain — put two +1/+1 counters on another target creature you control",
				twoPlusOneCountersOnFirstLegalTarget),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
