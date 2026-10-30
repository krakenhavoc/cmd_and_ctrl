package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Living Lightning, Charged Up — Legendary Creature — Elemental Hero {4}{R},
// 3/3 (Marvel Super Heroes Commander):
//
//	"Flying
//	 Haste
//	 At the beginning of combat on your turn, another target creature you
//	 control gets +1/+0 and gains haste until end of turn."
//
// Helitrooper's target-bound pump (firstTargetGainsUntilEndOfTurn) on a
// beginning-of-combat trigger. The target is "another" creature, so he never
// charges himself; the engine drops the trigger when there is no other
// creature of yours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b9397684-f226-4f77-a57a-88f97fbd96a5",
		Name:            "Living Lightning, Charged Up",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat("Living Lightning — another target creature you control gets +1/+0 and gains haste until end of turn",
				firstTargetGainsUntilEndOfTurn(1, "Living Lightning — +1/+0 and haste until end of turn", "haste")),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
