package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prophet of the Peak — Artifact Creature — Cat {6}, 5/5 (Throne of Eldraine, #5516):
//
//	"When this creature enters, scry 2."
//
// Automatic Librarian's shape: an enters trigger that scries two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eaa1634e-6ba1-41c2-91c0-967b36d13e14",
		Name:         "Prophet of the Peak",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Prophet of the Peak — scry 2", Do(Scry{N: 2})),
		},
	})
}
