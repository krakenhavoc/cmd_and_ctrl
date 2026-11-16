package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brotherhood Ambushers — Creature — Human Assassin {4}{B}, 6/3
// (Assassin's Creed, #5466):
//
//	"Freerunning {3}{B} (You may cast this spell for its freerunning
//	 cost if you dealt combat damage to a player this turn with an
//	 Assassin or commander.)"
//
// Freerunning is an alternative cost (CR 702.173), offered only once an
// Assassin or a commander you control has dealt combat damage to a
// player this turn; the Freerunning constructor carries that test.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "aea65f96-3c76-4761-9a7a-2d5c10d75ca0",
		Name:             "Brotherhood Ambushers",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Freerunning("{3}{B}", "Assassin")},
	})
}
