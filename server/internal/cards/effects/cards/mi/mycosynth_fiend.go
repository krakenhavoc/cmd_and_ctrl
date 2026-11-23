package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mycosynth Fiend — Creature — Phyrexian Horror {2}{G}, 2/2:
//
//	"This creature gets +1/+1 for each poison counter your opponents have."
//
// A layer 7c modify on this creature, summed over every opponent still in the
// game (the counters of all of them add up) and re-read on every recompute, so
// it follows the poison count as it rises.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7c4bc25-806c-4a75-85bf-a0e66308d5d8",
		Name:         "Mycosynth Fiend",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpSelfCreatureOrAttachedPer(1, 1, func(g *game.Game, source *game.Card) int {
				n := 0
				for _, p := range g.Seats {
					if p == nil || p.Eliminated || p.ID == source.Controller {
						continue
					}
					n += p.Poison
				}
				return n
			}),
		},
	})
}
