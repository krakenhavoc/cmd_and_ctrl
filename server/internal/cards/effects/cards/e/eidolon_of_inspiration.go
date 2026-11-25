package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eidolon of Inspiration — Enchantment Creature — Spirit {1}{W}{W}, 2/2
// (Theros Beyond Death, S82 OLD, #5514):
//
//	"At the beginning of combat on your turn, target creature you control gets
//	 +2/+0 until end of turn."
//
// The beginning-of-combat trigger of your own turn (not each combat) with a
// target creature you control, chosen as the trigger goes on the stack and
// pumped if still legal as it resolves (CR 608.2b); the Eidolon may target
// itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad70144e-16e4-4604-9e6d-cb287f67ddb5",
		Name:         "Eidolon of Inspiration",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.AtBeginningOfYourCombat("Eidolon of Inspiration — target creature you control gets +2/+0 until end of turn",
					effects.TrekB41PumpAndGrantEachTarget(2, 0, "Eidolon of Inspiration — +2/+0")),
				effects.TargetCreature("target creature you control", effects.YouControl())),
				effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0))),
		},
	})
}
