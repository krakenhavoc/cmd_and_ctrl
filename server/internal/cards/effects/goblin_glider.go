package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Glider — Creature — Goblin {1}{R}, 1/1 (Portal Second Age,
// #5625):
//
//	"Flying
//	 This creature can't block."
//
// Aesthir Glider's two lines: flying is printed data (CR 702.9b), and the
// block ban is a RestrictSelf static the declare-blockers check reads (CR
// 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "72bce821-46f0-40eb-abdb-575d4305b8fd",
		Name:            "Goblin Glider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})
}
