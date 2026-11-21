package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Young Wei Recruits — Creature — Human Soldier {1}{B}, 2/2
// (Portal Three Kingdoms, #5622):
//
//	"This creature can't block."
//
// "Can't block" is a static restriction on itself (RestrictSelf), read
// by the block check (CR 509.1b). It can still attack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cdada8f0-d001-485c-a245-5e1c1ab748ab",
		Name:         "Young Wei Recruits",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
