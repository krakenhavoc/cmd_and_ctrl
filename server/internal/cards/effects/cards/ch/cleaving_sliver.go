package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cleaving Sliver — Creature — Sliver {3}{R}, 2/2 (Modern Horizons, #5519):
//
//	"Sliver creatures you control get +2/+0."
//
// The shared tribal anthem in layer 7c, over your own Slivers only, the
// Cleaving Sliver included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b33059c-c862-4452-8609-14a0e2551fb3",
		Name:         "Cleaving Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, 2, 0),
		},
	})
}
