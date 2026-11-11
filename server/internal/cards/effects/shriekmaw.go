package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shriekmaw — Creature — Elemental {4}{B}, 3/2 (Murders at Karlov Manor
// Commander, #5368):
//
//	"Fear
//	 When this creature enters, destroy target nonartifact, nonblack
//	 creature.
//	 Evoke {1}{B} (You may cast this spell for its evoke cost. If you do,
//	 it's sacrificed when it enters.)"
//
// Ingot Chewer's shape with a creature clause: nonartifact and nonblack are
// two predicates on the target, read through the layers when it is chosen
// and again as the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "717bbb9d-0a5b-4979-a24b-9d54aac7657a",
		Name:            "Shriekmaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"fear"},
		AlternativeCosts: []game.AlternativeCost{
			Evoke("{1}{B}"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Shriekmaw — destroy target nonartifact, nonblack creature", destroyChosenPermanent),
				TargetCreature("target nonartifact, nonblack creature", Not(Artifact()), NonBlack())),
		},
	})
}
