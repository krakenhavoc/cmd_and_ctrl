package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plasma Elemental — Creature — Elemental {5}{U}, 4/1 (Fifth Dawn, #5604):
//
//	"This creature can't be blocked."
//
// Shoreline Looter's "can't be blocked" restriction on itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "419c501f-c107-4166-b56d-631672f2a5cc",
		Name:         "Plasma Elemental",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
	})
}
