package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lumengrid Drake — Creature — Drake {3}{U}, 2/2 (Scars of Mirrodin, S82 OLD
// batch 469, #5577):
//
//	"Flying
//	 Metalcraft — When this creature enters, if you control three or more
//	 artifacts, return target creature to its owner's hand."
//
// Rimekin Recluse's targeted enter-bounce (BounceChosenTarget) on the
// metalcraft gate: with fewer than three artifacts the trigger never goes on
// the stack, so no target is asked for (CR 603.4), and the count is checked
// again as it resolves. Any creature may be the target, the Drake itself
// included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0700d889-d235-49fe-8b92-698d486b2539",
		Name:            "Lumengrid Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.OldB469WhenThisEntersIfMetalcraft("Lumengrid Drake — return target creature to its owner's hand", effects.BounceChosenTarget),
				effects.TargetCreature("target creature"),
			),
		},
	})
}
