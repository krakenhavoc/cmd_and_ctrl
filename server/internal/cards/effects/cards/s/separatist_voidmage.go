package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Separatist Voidmage — Creature — Human Wizard {3}{U}, 2/2 (Magic Origins, #5549):
//
//	"When this creature enters, you may return target creature to its owner's hand."
//
// Man-o'-War's entry as a "you may": the creature is targeted as the
// trigger goes on the stack, and the question is asked as it resolves (CR
// 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9fabfa8-b930-428d-af4e-640fb310a5b2",
		Name:         "Separatist Voidmage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Separatist Voidmage — return target creature to its owner's hand", effects.BounceTheTargetToHand),
				effects.TargetCreature("target creature")),
				"Separatist Voidmage — return target creature to its owner's hand?"),
		},
	})
}
