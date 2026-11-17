package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hulking Cyclops — Creature — Cyclops {3}{R}{R}, 5/5 (Visions, #5631):
//
//	"This creature can't block."
//
// A static restriction on itself (RestrictSelf), read by the block check
// (CR 509.1b). It can still attack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04936918-7ef3-42fa-b0cd-87d18cb69d2c",
		Name:         "Hulking Cyclops",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})
}
