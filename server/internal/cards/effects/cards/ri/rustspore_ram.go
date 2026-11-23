package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rustspore Ram — Artifact Creature — Sheep {4}, 1/3 (Mirrodin, #5606):
//
//	"When this creature enters, destroy target Equipment."
//
// Viridian Shaman's enters trigger narrowed to an Equipment: the target
// is chosen as the trigger goes on the stack (CR 603.3d) and checked
// again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a45b3934-1c9b-4cff-98b1-c9ac2f7759ea",
		Name:         "Rustspore Ram",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Rustspore Ram — destroy target Equipment", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target Equipment", effects.HasSubtype("Equipment"))),
		},
	})
}
