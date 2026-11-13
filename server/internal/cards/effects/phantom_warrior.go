package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phantom Warrior — Creature — Illusion Warrior {1}{U}{U}, 2/2 (Portal, #5630):
//
//	"This creature can't be blocked."
//
// A static restriction on itself (RestrictSelf), read by the block check
// (CR 509.1b): no creature can block it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23745133-e5e2-4ce3-b94a-73d0d3d8a013",
		Name:         "Phantom Warrior",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
	})
}
