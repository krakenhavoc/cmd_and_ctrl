package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voidwielder — Creature — Human Wizard {4}{U}, 1/4 (Return to Ravnica, #5566):
//
//	"When this creature enters, you may return target creature to its owner's
//	 hand."
//
// The creature is chosen as the trigger goes on the stack (any creature, this one
// included) and the "may" is asked on resolution (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6686b897-64a9-4482-9563-57c88bc83d71",
		Name:         "Voidwielder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Voidwielder — you may return target creature to its owner's hand", effects.BounceTheTargetToHand),
				effects.TargetCreature("target creature")),
				"Voidwielder — return target creature to its owner's hand?"),
		},
	})
}
