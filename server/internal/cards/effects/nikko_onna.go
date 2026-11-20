package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nikko-Onna — Creature — Spirit {2}{W}, 2/2 (Saviors of Kamigawa, #5601):
//
//	"When this creature enters, destroy target enchantment.
//	 Whenever you cast a Spirit or Arcane spell, you may return this
//	 creature to its owner's hand."
//
// The enters trigger uses the stack. A target that left in response is skipped (CR 608.2b). The return is the cycle's
// shared sokB1OnnaReturn: optional, asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5cf3f74-e497-46d0-acfb-3b57ee27a314",
		Name:         "Nikko-Onna",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Nikko-Onna — destroy target enchantment", destroyFirstLegalCardTarget),
				TargetPermanent("target enchantment", Enchantment())),
			sokB1OnnaReturn("Nikko-Onna"),
		},
	})
}
