package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bonesplitter Sliver — Creature — Sliver {3}{R}, 2/2 (Time Spiral, #5596):
//
//	"All Sliver creatures get +2/+0."
//
// Blade Sliver's anthem at +2/+0: every Sliver creature at the table, this
// one and an opponent's included, in layer 7c (CR 613.4c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a0f6834c-faa9-4d90-adad-847a8b337fe6",
		Name:         "Bonesplitter Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Sliver"}}, 2, 0),
		},
	})
}
