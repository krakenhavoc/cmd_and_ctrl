package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kyren Legate — Creature — Goblin {1}{R}, 1/1 (Mercadian Masques, #5619):
//
//	"If an opponent controls a Plains and you control a Mountain, you may
//	 cast this spell without paying its mana cost.
//	 Haste"
//
// The Legate cycle's free cast (FreeIfAnOpponentControlsAAndYouControlA):
// an OPPONENT controls a Plains and you control a Mountain, read off
// effective subtypes as you cast it. Haste is a printed keyword, so it
// can attack the turn it enters either way it was cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "41b46aa5-6091-4bc4-8d17-d2c6b5fe1c5e",
		Name:             "Kyren Legate",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"haste"},
		AlternativeCosts: []game.AlternativeCost{FreeIfAnOpponentControlsAAndYouControlA("Plains", "Mountain")},
	})
}
