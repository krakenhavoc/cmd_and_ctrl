package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e09661d4-858b-43a8-b6a4-ba76f7a69777",
		Name:         "Kiri-Onna",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Kiri-Onna — return target creature to its owner's hand", bounceTheTargetToHand),
				TargetCreature("target creature")),
			sokB1OnnaReturn("Kiri-Onna"),
		},
	})
}
