package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a45b3934-1c9b-4cff-98b1-c9ac2f7759ea",
		Name:         "Rustspore Ram",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Rustspore Ram — destroy target Equipment", destroyFirstLegalCardTarget),
				TargetPermanent("target Equipment", HasSubtype("Equipment"))),
		},
	})
}
