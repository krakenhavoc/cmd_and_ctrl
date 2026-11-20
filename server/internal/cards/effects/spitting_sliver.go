package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spitting Sliver — Creature — Sliver {4}{B}, 3/3 (Planar Chaos, #5595):
//
//	"All Sliver creatures have first strike."
//
// A layer-6 keyword grant (CR 613.1f) of first strike to every Sliver creature at
// the table, an opponent's and this one included (Fury Sliver's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8415a482-3957-4c00-9015-99defc049877",
		Name:         "Spitting Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}}, "first strike"),
		},
	})
}
