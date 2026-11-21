package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tormented Soul — Creature — Spirit {B}, 1/1 (Magic 2012, #5573):
//
//	"This creature can't block and can't be blocked."
//
// Marang River Prowler's two self restrictions, can't block and can't be
// blocked (RestrictSelf).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee334bd7-d344-4684-85c8-d0464bbaf2b4",
		Name:         "Tormented Soul",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.RestrictSelf(game.CantBlock),
			effects.RestrictSelf(game.CantBeBlocked),
		},
	})
}
