package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spineless Thug — Creature — Phyrexian Zombie Mercenary {1}{B}, 2/2
// (Nemesis, #5618):
//
//	"This creature can't block."
//
// Sightless Ghoul's restriction without undying: RestrictSelf with
// game.CantBlock.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78bbd213-e8f4-40da-aef5-853244e95756",
		Name:         "Spineless Thug",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})
}
