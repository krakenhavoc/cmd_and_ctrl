package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yuki-Onna — Creature — Spirit {3}{R}, 3/1 (Saviors of Kamigawa, #5601):
//
//	"When this creature enters, destroy target artifact.
//	 Whenever you cast a Spirit or Arcane spell, you may return this
//	 creature to its owner's hand."
//
// The enters trigger uses the stack. A target that left in response is skipped (CR 608.2b). The return is the cycle's
// shared sokB1OnnaReturn: optional, asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f867a2c-f17b-4608-bca7-5b49d1387bc7",
		Name:         "Yuki-Onna",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Yuki-Onna — destroy target artifact", destroyFirstLegalCardTarget),
				TargetPermanent("target artifact", Artifact())),
			sokB1OnnaReturn("Yuki-Onna"),
		},
	})
}
