package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian Automaton — Artifact Creature — Construct {4}, 3/3 (Magic Origins, #5549):
//
//	"When this creature dies, you gain 3 life."
//
// Anodet Lurker's death trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd29d0fd-685f-4c26-b69c-e96f1c0d59c4",
		Name:         "Guardian Automaton",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Guardian Automaton — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
