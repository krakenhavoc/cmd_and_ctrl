package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yellow Scarves General — Creature — Human Soldier {3}{R}, 2/2
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
	Register(Spec{
		OracleID:        "6fbade4d-502d-45d7-a79e-3fd9e3e7ccdf",
		Name:            "Yellow Scarves General",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})
}
