package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sinew Sliver — Creature — Sliver {1}{W}, 1/1 (Commander Masters, #5948):
//
//	"All Sliver creatures get +1/+1."
//
// A tribal anthem with neither "other" nor "you control" (layer 7c,
// CR 613.4c): every Sliver at the table, this one and an opponent's
// included.
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
