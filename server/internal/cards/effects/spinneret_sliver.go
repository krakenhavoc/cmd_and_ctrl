package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spinneret Sliver — Creature — Sliver {1}{G}, 2/2 (Time Spiral, #5596):
//
//	"All Sliver creatures have reach."
//
// A layer-6 keyword grant (CR 613.1f) of reach (CR 702.17a) to every
// Sliver creature at the table, an opponent's included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fbdec63e-0679-47e9-b82a-daa56df50481",
		Name:         "Spinneret Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}}, "reach"),
		},
	})
}
