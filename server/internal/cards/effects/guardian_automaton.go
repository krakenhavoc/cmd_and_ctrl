package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guardian Automaton — Artifact Creature — Construct {4}, 3/3 (Magic Origins, #5549):
//
//	"When this creature dies, you gain 3 life."
//
// Anodet Lurker's death trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cd29d0fd-685f-4c26-b69c-e96f1c0d59c4",
		Name:         "Guardian Automaton",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Guardian Automaton — you gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
