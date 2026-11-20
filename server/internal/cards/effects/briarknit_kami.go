package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Briarknit Kami — Creature — Spirit {3}{G}{G}, 3/3 (Saviors of Kamigawa,
// #5601):
//
//	"Whenever you cast a Spirit or Arcane spell, put a +1/+1 counter on
//	 target creature."
//
// The ki cycle's cast trigger (chkB1SpiritOrArcane) with Hunting Moa's
// counter; the creature is chosen as the trigger goes on the stack, and
// one that left in response gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "913dcf97-0804-4b21-8daa-f2b481dc0f2f",
		Name:         "Briarknit Kami",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverYouCast(chkB1SpiritOrArcane(), "Briarknit Kami — put a +1/+1 counter on target creature",
				putPlusOneCounterOnEachLegalTarget), TargetCreature("target creature")),
		},
	})
}
