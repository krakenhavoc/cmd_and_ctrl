package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "913dcf97-0804-4b21-8daa-f2b481dc0f2f",
		Name:         "Briarknit Kami",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverYouCast(effects.ChkB1SpiritOrArcane(), "Briarknit Kami — put a +1/+1 counter on target creature",
				effects.PutPlusOneCounterOnEachLegalTarget), effects.TargetCreature("target creature")),
		},
	})
}
