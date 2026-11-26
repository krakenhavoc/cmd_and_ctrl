package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keeper of the Lens — Artifact Creature — Golem {1}, 1/2 (Dragons of Tarkir):
//
//	"You may look at face-down creatures you don't control any time."
//
// The standing look (LooksAtFaceDownCreatures): derived from the battlefield on
// every frame, never written into the creatures' knowers, and gone the moment
// the Keeper leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                 "84b28100-ad70-4a78-b557-99f5ef6126a3",
		Name:                     "Keeper of the Lens",
		Completeness:             effects.CompletenessFull,
		LooksAtFaceDownCreatures: true,
	})
}
