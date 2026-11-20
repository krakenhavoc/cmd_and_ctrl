package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reflex Sliver — Creature — Sliver {3}{G}, 2/2 (Planar Chaos, #5595):
//
//	"All Sliver creatures have haste."
//
// A layer-6 keyword grant (CR 613.1f) of haste to every Sliver creature at
// the table, an opponent's and this one included (Fury Sliver's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "80e4e857-3f7e-4344-bf41-64d3b32f328f",
		Name:         "Reflex Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}}, "haste"),
		},
	})
}
