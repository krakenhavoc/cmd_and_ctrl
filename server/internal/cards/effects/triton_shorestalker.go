package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Triton Shorestalker — Creature — Merfolk Rogue {U}, 1/1:
//
//	"This creature can't be blocked."
//
// The S24 restriction bit on the creature itself, read at the block
// declaration (CR 509.1b). No simplification.
func init() {
	Register(Spec{
		OracleID:     "71a60247-17e5-44bd-9fed-f7d550b09ef3",
		Name:         "Triton Shorestalker",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
	})
}
