package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Automatic Librarian — Artifact Creature — Construct {3}, 3/2 (Dominaria United, S82):
//
//	"When this creature enters, scry 2."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb570331-35da-44b8-b711-ccf7444d8e39",
		Name:         "Automatic Librarian",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Automatic Librarian — scry 2", Do(Scry{N: 2})),
		},
	})
}
