package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enatu Golem — Artifact Creature — Golem {6}, 3/5 (Rise of the Eldrazi,
// S82 OLD batch 481, #5581):
//
//	"When this creature dies, you gain 4 life."
//
// A fixed gain by the controller as the Golem's own dies trigger resolves
// (the controller is read from the dying permanent, CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1cb3f27c-2670-43b8-9305-8adf2e7bd3ae",
		Name:         "Enatu Golem",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Enatu Golem — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}
