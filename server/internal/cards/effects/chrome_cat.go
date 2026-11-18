package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chrome Cat (Streets of New Capenna, #5491):
//
//	When this creature enters, scry 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0e2f8758-7457-4b54-ae61-a9c68cea4a23",
		Name:         "Chrome Cat",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Chrome Cat — scry 1", Do(Scry{N: 1})),
		},
	})
}
