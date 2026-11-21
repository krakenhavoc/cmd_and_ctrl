package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "8f867a2c-f17b-4608-bca7-5b49d1387bc7",
		Name:         "Yuki-Onna",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Yuki-Onna — destroy target artifact", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target artifact", effects.Artifact())),
			effects.SokB1OnnaReturn("Yuki-Onna"),
		},
	})
}
