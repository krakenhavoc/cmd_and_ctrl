package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "598f948b-bdd3-490f-b3e9-f0f9dc470522",
		Name:         "Onulet",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Onulet — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
