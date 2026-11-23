package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chrome Cat (Streets of New Capenna, #5491):
//
//	When this creature enters, scry 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e2f8758-7457-4b54-ae61-a9c68cea4a23",
		Name:         "Chrome Cat",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Chrome Cat — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
