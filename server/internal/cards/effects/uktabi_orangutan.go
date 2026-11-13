package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6cc07a7e-e70c-45c5-83c5-a2704e57dbc5",
		Name:         "Uktabi Orangutan",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Uktabi Orangutan — destroy target artifact", destroyFirstLegalCardTarget),
				TargetPermanent("target artifact", Artifact())),
		},
	})
}
