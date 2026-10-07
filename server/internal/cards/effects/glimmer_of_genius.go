package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glimmer of Genius — Instant {3}{U}:
//
//	"Scry 2, then draw two cards. You get {E}{E} (two energy counters)."
//
// ADR 0129 PR 1. The draw and the energy follow the scry prompt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9c58cfb-b7d5-4e86-a5d3-efb07192aa22",
		Name:         "Glimmer of Genius",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Energy: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			return Scry{Player: controller, N: 2, Then: func(g *game.Game) error {
				if err := g.DrawNForEffect(controller, 2); err != nil {
					return err
				}
				return g.AddPlayerCounterByForEffect(controller, controller, game.CounterEnergy, 2)
			}}.Apply(ctx)
		},
	})
}
