package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wild Celebrants — Creature — Satyr {3}{R}{R}, 5/3 (Commander Legends):
//
//	"When this creature enters, you may destroy target artifact."
//
// War Priest of Thune's optional targeted enters trigger over an artifact:
// the target is chosen as the trigger goes on the stack and the "you may" is
// asked as it resolves, so the controller can spare their own artifact.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "695661ac-6f50-4a0e-9831-a0e285bab4bf",
		Name:         "Wild Celebrants",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Wild Celebrants — destroy target artifact", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target artifact", effects.Artifact())),
				"Wild Celebrants — destroy target artifact?"),
		},
	})
}
