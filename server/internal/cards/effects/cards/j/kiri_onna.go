package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kiri-Onna — Creature — Spirit {4}{U}, 2/2 (Saviors of Kamigawa, #5601):
//
//	"When this creature enters, return target creature to its owner's hand.
//	 Whenever you cast a Spirit or Arcane spell, you may return this
//	 creature to its owner's hand."
//
// The enters trigger uses the stack. It is Man-o'-War's, and may target the Onna itself. The return is the cycle's
// shared sokB1OnnaReturn: optional, asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e09661d4-858b-43a8-b6a4-ba76f7a69777",
		Name:         "Kiri-Onna",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Kiri-Onna — return target creature to its owner's hand", effects.BounceTheTargetToHand),
				effects.TargetCreature("target creature")),
			effects.SokB1OnnaReturn("Kiri-Onna"),
		},
	})
}
