package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uktabi Orangutan — Creature — Ape {2}{G}, 2/2 (Visions, #5631):
//
//	"When this creature enters, destroy target artifact."
//
// Keldon Vandals' trigger without the echo: it targets as it goes on the
// stack (CR 603.3d), and a target that has left does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6cc07a7e-e70c-45c5-83c5-a2704e57dbc5",
		Name:         "Uktabi Orangutan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Uktabi Orangutan — destroy target artifact", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target artifact", effects.Artifact())),
		},
	})
}
