package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tormented Soul — Creature — Spirit {B}, 1/1 (Magic 2012, #5573):
//
//	"This creature can't block and can't be blocked."
//
// Marang River Prowler's two self restrictions, can't block and can't be
// blocked (RestrictSelf).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ee334bd7-d344-4684-85c8-d0464bbaf2b4",
		Name:         "Tormented Soul",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			RestrictSelf(game.CantBlock),
			RestrictSelf(game.CantBeBlocked),
		},
	})
}
