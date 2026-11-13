package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hulking Goblin — Creature — Goblin {1}{R}, 2/2 (Portal, #5630):
//
//	"This creature can't block."
//
// A static restriction on itself (RestrictSelf), read by the block check
// (CR 509.1b). It can still attack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1451b623-8214-4d81-9436-62dbe8b79539",
		Name:         "Hulking Goblin",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})
}
