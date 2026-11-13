package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Metathran Soldier — Creature — Metathran Soldier {1}{U}, 1/1 (Urza's
// Destiny, #5621):
//
//	"This creature can't be blocked."
//
// Slither Blade's restriction on a 1/1: the declare-blockers check
// refuses any block of it (CR 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b723fcec-f208-4397-a178-e1e47b729cf1",
		Name:         "Metathran Soldier",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
	})
}
