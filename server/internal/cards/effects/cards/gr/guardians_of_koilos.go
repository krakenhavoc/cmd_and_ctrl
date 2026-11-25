package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guardians of Koilos — Artifact Creature — Construct {5}, 4/4 (Dominaria,
// S82 OLD batch 214, #5530):
//
//	"When this creature enters, you may return another target historic
//	 permanent you control to its owner's hand. (Artifacts, legendaries, and
//	 Sagas are historic.)"
//
// Jeskai Barricade's optional bounce with a historic target. "Another"
// excludes the Guardians by object (it is an artifact, so it is historic
// itself), the target is re-checked as the trigger resolves (CR 608.2b), and
// with no other historic permanent you control the trigger is removed
// (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f593bed-4be8-403d-bce2-fb769658e673",
		Name:         "Guardians of Koilos",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.WhenThisEnters("Guardians of Koilos — return another target historic permanent you control to its owner's hand", effects.BounceChosenTarget),
					"Guardians of Koilos — return another historic permanent you control to its owner's hand?"),
				effects.Another(effects.TargetPermanent("another target historic permanent you control", effects.YouControl(), effects.B14Historic()))),
		},
	})
}
