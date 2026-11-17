package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wandering Eye — Creature — Illusion {2}{U}, 1/3
// (Nemesis, #5618):
//
//	"Flying
//	 Players play with their hands revealed."
//
// Flying is printed data. "Players play with their hands revealed" is
// Revelation's HandsRevealed scope, every player's hand, for as long as
// the Eye is on the battlefield (ADR 0199).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "59015068-6868-4235-981b-96137123c685",
		Name:            "Wandering Eye",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		HandsRevealed:   game.HandRevealEachPlayer,
	})
}
