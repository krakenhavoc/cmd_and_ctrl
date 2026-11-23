package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "92f5f80c-de1a-44e8-93f1-e405241c1e82",
		Name:         "Might Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Sliver"}}, 2, 2),
		},
	})
}
