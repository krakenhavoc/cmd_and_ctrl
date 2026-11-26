package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dispersal Technician — Creature — Vedalken Artificer {4}{U}, 3/2 (Aether
// Revolt, S82 older sets, #5536):
//
//	"When this creature enters, you may return target artifact to its owner's
//	 hand."
//
// The target (any artifact, yours or not, the Technician's own controller's
// included) is chosen as the trigger goes on the stack and the "may" is
// answered as it resolves; a target that is gone does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "47a66451-8a4e-4fc6-8685-96e1feb8f997",
		Name:         "Dispersal Technician",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Dispersal Technician — you may return target artifact to its owner's hand",
				effects.BounceTheTargetToHand),
				effects.TargetPermanent("target artifact", effects.Artifact())),
				"Dispersal Technician — return target artifact to its owner's hand?"),
		},
	})
}
