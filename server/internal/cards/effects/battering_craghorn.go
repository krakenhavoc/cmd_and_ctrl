package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battering Craghorn — Creature — Goat Beast {2}{R}{R}, 3/1 (Onslaught):
//
//	"First strike
//	 Morph {1}{R}{R}"
//
// Ainok Tracker's shape: face down it is a 2/2 with no first strike
// (CR 708.2a); turned face up for {1}{R}{R} it is the 3/1 first striker,
// the same object (CR 708.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "a084bca4-fe72-443e-adf4-ab5456c7e699",
		Name:             "Battering Craghorn",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"first strike"},
		AlternativeCosts: []game.AlternativeCost{Morph("{1}{R}{R}")},
	})
}
