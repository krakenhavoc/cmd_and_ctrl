package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oxidda Scrapmelter — Creature — Beast {3}{R}, 3/3 (Scars of Mirrodin, #5577):
//
//	"When this creature enters, destroy target artifact."
//
// Uktabi Orangutan's enters trigger: the target is chosen as it goes on
// the stack (CR 603.3d) and destroyed if still legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cef64ec6-cdfb-42a4-9a4a-3ef5448e93fc",
		Name:         "Oxidda Scrapmelter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Oxidda Scrapmelter — destroy target artifact", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target artifact", effects.Artifact())),
		},
	})
}
