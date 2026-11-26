package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peregrination — Sorcery {3}{G} (Born of the Gods, S82 OLD b378, #5559):
//
//	"Search your library for up to two basic land cards, reveal those cards,
//	 and put one onto the battlefield tapped and the other into your hand.
//	 Shuffle, then scry 1."
//
// Kodama's Reach's search (one basic onto the battlefield tapped, the other
// into hand, then the shuffle), then the scry as a second step, which waits
// for the search's prompts (CR 608.2c). The scry is a fixed rider, declared in
// Purpose.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2aff9be5-f6d4-4cdd-ab59-20202bc2f144",
		Name:         "Peregrination",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1, Tutors: 1, Scry: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.TwoLandSearch(ctx, "Peregrination", "basic land", effects.IsBasicLand, effects.S82OldB378PeregrinationHandTail)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
		),
	})
}
