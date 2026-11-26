package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Batterhorn — Creature — Beast {4}{R}, 4/3 (Return to Ravnica, #5566):
//
//	"When this creature enters, you may destroy target artifact."
//
// The artifact is chosen as the trigger goes on the stack and the "may" is
// asked on resolution (War Priest of Thune's shape); with no artifact on the
// battlefield the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9dd5841-7e4b-40bf-9969-e8fb5df46763",
		Name:         "Batterhorn",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Batterhorn — you may destroy target artifact", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target artifact", effects.Artifact())),
				"Batterhorn — destroy target artifact?"),
		},
	})
}
