package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surrak, the Hunt Caller — Legendary Creature — Human Warrior {2}{G}{G},
// 5/4 (Foundations, #4465):
//
//	"Formidable — At the beginning of combat on your turn, if creatures
//	 you control have total power 8 or greater, target creature you
//	 control gains haste until end of turn."
//
// Formidable is an intervening "if" (CR 603.4): the total is Tactical
// Officer's sum (current power after the layers, every creature you
// control), read as the beginning of combat step begins and again as the
// ability resolves. The target may be Surrak itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa56a5fa-ef96-404c-8fb6-d1f5fcebb52e",
		Name:         "Surrak, the Hunt Caller",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.IfVentureCond(effects.VentureCond(effects.CreaturesYouControlTotalPowerAtLeast(8)),
				effects.TriggerWithPurpose(effects.Targeting(
					effects.AtBeginningOfYourCombat("Surrak, the Hunt Caller — formidable: target creature you control gains haste until end of turn",
						effects.JmpB7FirstLegalTargetGains("Surrak, the Hunt Caller — haste", "haste")),
					effects.TargetCreature("target creature you control", effects.YouControl())),
					effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")))),
		},
	})
}
