package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lost Legion — Creature — Spirit Knight {1}{B}{B}, 2/3 (Throne of Eldraine, #5516):
//
//	"When this creature enters, scry 2."
//
// Automatic Librarian's shape: an enters trigger that scries two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b8f3b5b5-8835-400f-a626-94ff0a007e56",
		Name:         "Lost Legion",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lost Legion — scry 2", Do(Scry{N: 2})),
		},
	})
}
