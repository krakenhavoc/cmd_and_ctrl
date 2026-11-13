package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cho-Arrim Legate — Creature — Human Soldier {2}{W}, 1/2
// (Mercadian Masques, #5619):
//
//	"If an opponent controls a Swamp and you control a Plains, you may
//	 cast this spell without paying its mana cost.
//	 Protection from black"
//
// The Legate cycle's free cast (FreeIfAnOpponentControlsAAndYouControlA):
// an OPPONENT controls a Swamp and you control a Plains, read off
// effective subtypes as you cast it. Protection from black is a printed
// keyword (CR 702.16).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "780c9cb2-9d98-4789-87e0-cb81db477204",
		Name:             "Cho-Arrim Legate",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"protection from black"},
		AlternativeCosts: []game.AlternativeCost{FreeIfAnOpponentControlsAAndYouControlA("Swamp", "Plains")},
	})
}
