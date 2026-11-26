package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lens of Clarity — Artifact {1} (Khans of Tarkir):
//
//	"You may look at the top card of your library and at face-down creatures you
//	 don't control any time."
//
// Two standing looks, each derived from the battlefield on every frame and
// never written into a card's knowers: Muad'Dib's look at your own top card
// (LibraryTopOwner, private to you) and Keeper of the Lens's look at
// face-down creatures you don't control. Both end the moment the Lens leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                 "7df50107-25d8-47dc-9a13-27b257cfbf6f",
		Name:                     "Lens of Clarity",
		Completeness:             effects.CompletenessFull,
		LibraryTopVisible:        game.LibraryTopOwner,
		LooksAtFaceDownCreatures: true,
	})
}
