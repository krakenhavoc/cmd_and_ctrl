package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leveler — Artifact Creature — Juggernaut {5}, 10/10 (Secret Lair Drop):
//
//	"When this creature enters, exile all cards from your library."
//
// One exile of every card in the controller's library, none kept and none
// looked at. An empty library exiles nothing, and exiling is not drawing, so
// it is not a loss on its own (CR 704.5b needs a draw).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ba1d34ab-7529-4f47-a5e7-7d8f66d7b901",
		Name:         "Leveler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Leveler — exile all cards from your library", effects.S82OldB023ExileYourLibrary),
		},
	})
}
