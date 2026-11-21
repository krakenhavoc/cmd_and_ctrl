package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viridian Shaman — Creature — Elf Shaman {2}{G}, 2/2 (Mirrodin, #5606):
//
//	"When this creature enters, destroy target artifact."
//
// Uktabi Orangutan's enters trigger: the target is chosen as the
// trigger goes on the stack (CR 603.3d) and checked again as it
// resolves (CR 608.2b); with no artifact on the battlefield the
// trigger is removed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5aa6d553-a144-4c5b-83fa-65e931903899",
		Name:         "Viridian Shaman",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Viridian Shaman — destroy target artifact", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target artifact", effects.Artifact())),
		},
	})
}
