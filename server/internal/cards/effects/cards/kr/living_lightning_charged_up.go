package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "b9397684-f226-4f77-a57a-88f97fbd96a5",
		Name:            "Living Lightning, Charged Up",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat("Living Lightning — another target creature you control gets +1/+0 and gains haste until end of turn",
				effects.FirstTargetGainsUntilEndOfTurn(1, "Living Lightning — +1/+0 and haste until end of turn", "haste")),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
