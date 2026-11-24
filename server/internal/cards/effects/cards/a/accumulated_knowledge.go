package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Accumulated Knowledge — Instant {1}{U} (Secret Lair Drop):
//
//	"Draw a card, then draw cards equal to the number of cards named
//	 Accumulated Knowledge in all graveyards."
//
// The count is taken after the first draw, as the spell resolves; the spell
// itself is still on the stack, so it is not counted (it goes to the graveyard
// only once it has finished resolving).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ca25379-d194-4670-a7bf-525ad11cf436",
		Name:         "Accumulated Knowledge",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 1},
			effects.DrawForEach{Name: "cards named Accumulated Knowledge in all graveyards", Count: accumulatedKnowledgeInGraveyards},
		),
	})
}

func accumulatedKnowledgeInGraveyards(ctx *effects.Context) int {
	n := 0
	for _, p := range ctx.Game.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.Name == "Accumulated Knowledge" {
				n++
			}
		}
	}
	return n
}
