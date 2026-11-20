package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Might Sliver — Creature — Sliver {4}{G}, 2/2 (Time Spiral, #5596):
//
//	"All Sliver creatures get +2/+2."
//
// A tribal anthem with neither "other" nor "you control": every Sliver
// creature at the table gets +2/+2, this one and an opponent's included.
// Layer 7c (CR 613.4c), so it stacks with the other Sliver lords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92f5f80c-de1a-44e8-93f1-e405241c1e82",
		Name:         "Might Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Sliver"}}, 2, 2),
		},
	})
}
