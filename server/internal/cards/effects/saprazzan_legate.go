package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Saprazzan Legate — Creature — Merfolk Soldier {3}{U}, 1/3
// (Mercadian Masques, #5619):
//
//	"If an opponent controls a Mountain and you control an Island, you
//	 may cast this spell without paying its mana cost.
//	 Flying"
//
// The Legate cycle's free cast (FreeIfAnOpponentControlsAAndYouControlA,
// Sivvi's Ruse's condition): an OPPONENT controls a Mountain and you
// control an Island, read off effective subtypes as you cast it. Flying
// is a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d20b62fc-e248-4c84-87ac-7319b92531d4",
		Name:             "Saprazzan Legate",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{FreeIfAnOpponentControlsAAndYouControlA("Mountain", "Island")},
	})
}
