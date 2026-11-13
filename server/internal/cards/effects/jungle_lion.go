package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jungle Lion — Creature — Cat {G}, 2/1 (Portal, #5630):
//
//	"This creature can't block."
//
// A static restriction on itself (RestrictSelf), read by the block check
// (CR 509.1b). It can still attack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "45b42694-9ee8-44f3-9de3-aa2c0f1b6a13",
		Name:         "Jungle Lion",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})
}
