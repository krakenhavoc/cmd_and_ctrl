package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Craven Knight — Creature — Human Knight {1}{B}, 2/2 (Portal, #5630):
//
//	"This creature can't block."
//
// A static restriction on itself (RestrictSelf), read by the block check
// (CR 509.1b). It can still attack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "27d138af-7c5d-4c7e-a310-63d5b03f039d",
		Name:         "Craven Knight",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})
}
