package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "717bbb9d-0a5b-4979-a24b-9d54aac7657a",
		Name:            "Shriekmaw",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"fear"},
		AlternativeCosts: []game.AlternativeCost{
			effects.Evoke("{1}{B}"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Shriekmaw — destroy target nonartifact, nonblack creature", effects.DestroyChosenPermanent),
				effects.TargetCreature("target nonartifact, nonblack creature", effects.Not(effects.Artifact()), effects.NonBlack())),
		},
	})
}
