package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yellow Scarves Troops — Creature — Human Soldier {1}{R}, 2/2
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
		OracleID:     "fcfa5671-a7b1-4170-9c72-f18a9eccfa31",
		Name:         "Yellow Scarves Troops",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
