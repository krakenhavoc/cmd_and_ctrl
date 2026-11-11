package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ingot Chewer — Creature — Elemental {4}{R}, 3/3 (Lorwyn Eclipsed
// Commander, #5368):
//
//	"When this creature enters, destroy target artifact.
//	 Evoke {R} (You may cast this spell for its evoke cost. If you do,
//	 it's sacrificed when it enters.)"
//
// Mulldrifter's evoke (CR 702.74a) on Avalanche Riders' targeted destroy:
// the evoke sacrifice is a separate trigger, so the destroy trigger is
// put on the stack with it and resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "127795e6-3f0e-4318-95ad-ccb1912580ac",
		Name:         "Ingot Chewer",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Evoke("{R}"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Ingot Chewer — destroy target artifact", destroyChosenPermanent),
				TargetPermanent("target artifact", Artifact())),
		},
	})
}
