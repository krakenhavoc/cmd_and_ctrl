package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sinew Sliver — Creature — Sliver {1}{W}, 1/1 (Planar Chaos, #5595):
//
//	"All Sliver creatures get +1/+1."
//
// Might Sliver's tribal anthem at +1/+1: every Sliver creature at the
// table, this one and an opponent's included. Layer 7c (CR 613.4c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ffff90c3-63c4-4dee-a21d-6b2b113f4f80",
		Name:         "Sinew Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Sliver"}}, 1, 1),
		},
	})
}
