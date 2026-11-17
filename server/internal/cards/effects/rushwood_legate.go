package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rushwood Legate — Creature — Dryad {2}{G}, 2/1 (Mercadian Masques, #5619):
//
//	"If an opponent controls an Island and you control a Forest, you may
//	 cast this spell without paying its mana cost."
//
// The Legate cycle's free cast (FreeIfAnOpponentControlsAAndYouControlA):
// an OPPONENT controls an Island and you control a Forest, read off
// effective subtypes as you cast it. The Dryad has no other text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f3623759-bf3a-4846-9ef7-1b732ad4355f",
		Name:             "Rushwood Legate",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{FreeIfAnOpponentControlsAAndYouControlA("Island", "Forest")},
	})
}
