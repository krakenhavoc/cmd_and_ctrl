package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "fa56a5fa-ef96-404c-8fb6-d1f5fcebb52e",
		Name:         "Surrak, the Hunt Caller",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			IfVentureCond(ventureCond(creaturesYouControlTotalPowerAtLeast(8)),
				TriggerWithPurpose(Targeting(
					AtBeginningOfYourCombat("Surrak, the Hunt Caller — formidable: target creature you control gains haste until end of turn",
						jmpB7FirstLegalTargetGains("Surrak, the Hunt Caller — haste", "haste")),
					TargetCreature("target creature you control", YouControl())),
					ForTargets(BenefitUntilEOT(0, 0, 0, "haste")))),
		},
	})
}
