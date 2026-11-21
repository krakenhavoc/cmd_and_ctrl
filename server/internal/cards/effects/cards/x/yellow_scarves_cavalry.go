package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yellow Scarves Cavalry — Creature — Human Soldier {1}{R}, 1/1
// (Portal Three Kingdoms, #5622):
//
//	"Horsemanship (This creature can't be blocked except by creatures with horsemanship.)
//	 This creature can't block."
//
// Horsemanship is an engine keyword carried as printed data: the block
// check refuses any blocker without horsemanship. It doesn't stop this
// creature blocking.
//
// "Can't block" is a static restriction on itself (RestrictSelf), read
// by the block check (CR 509.1b). It can still attack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ff7e8e9-11b4-4a77-a423-70777f00482b",
		Name:            "Yellow Scarves Cavalry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
