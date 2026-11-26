package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Midvast Protector — Creature — Human Wizard {3}{W}, 2/3 (Avacyn Restored,
// S82 OLD, #5570):
//
//	"When this creature enters, target creature you control gains protection
//	 from the color of your choice until end of turn."
//
// Benevolent Bodyguard's effect on an enter trigger: the creature is targeted
// as the trigger goes on the stack and the colour is chosen as it resolves,
// as the colour the creature is protected FROM. The Protector is a legal
// target itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33a5ad42-cd81-4724-9f68-7f32a1c8a047",
		Name:         "Midvast Protector",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEnters("Midvast Protector — target creature you control gains protection from the color of your choice until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ProtectionFromAChosenColorForTheTarget("Midvast Protector")(item, effects.NewContext(g, item))
					}),
				effects.TargetCreature("target creature you control", effects.YouControl()),
			), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection"))),
		},
	})
}
