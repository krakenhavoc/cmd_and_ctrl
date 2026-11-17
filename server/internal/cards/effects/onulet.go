package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Onulet — Artifact Creature — Construct {3}, 2/2 (Antiquities, #5640):
//
//	"When this creature dies, you gain 2 life."
//
// Curious Farm Animals' death trigger for two: it goes on the stack when
// the Onulet reaches the graveyard from the battlefield, and its
// controller gains the life as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "598f948b-bdd3-490f-b3e9-f0f9dc470522",
		Name:         "Onulet",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Onulet — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
