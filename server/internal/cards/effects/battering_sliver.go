package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battering Sliver — Creature — Sliver {5}{R}, 4/4 (Planar Chaos, #5595):
//
//	"All Sliver creatures have trample."
//
// A layer-6 keyword grant (CR 613.1f) of trample to every Sliver creature at
// the table, an opponent's and this one included (Fury Sliver's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4adf959b-2e09-47ff-a1b1-b889348735b8",
		Name:         "Battering Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}}, "trample"),
		},
	})
}
