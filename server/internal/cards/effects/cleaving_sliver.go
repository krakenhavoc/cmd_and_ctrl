package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cleaving Sliver — Creature — Sliver {3}{R}, 2/2 (Modern Horizons, #5519):
//
//	"Sliver creatures you control get +2/+0."
//
// The shared tribal anthem in layer 7c, over your own Slivers only, the
// Cleaving Sliver included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9b33059c-c862-4452-8609-14a0e2551fb3",
		Name:         "Cleaving Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, 2, 0),
		},
	})
}
