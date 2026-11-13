package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anodet Lurker — Artifact Creature — Construct {5}, 3/3 (Fifth Dawn,
// #5604):
//
//	"When this creature dies, you gain 3 life."
//
// Onulet's death trigger for three.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5c2dbe90-5784-4bc4-a715-76a61c1ecfbc",
		Name:         "Anodet Lurker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Anodet Lurker — you gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
