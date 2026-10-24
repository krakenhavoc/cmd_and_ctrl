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
			return Scry{Player: ctx.Controller(), N: 2, ThenTail: ctx.ItemTail(glimmerOfGeniusThen)}.Apply(ctx)
		},
	})
}

// glimmerOfGeniusThen is the draws and the energy, once the scry is done
// (#4253 slice 4). An on-disk key.
var glimmerOfGeniusThen = CardTail("glimmer-of-genius/then", func(ctx *Context, _ game.TailOutcome) error {
	return glimmerOfGeniusRest.Run(ctx)
})

// glimmerOfGeniusRest is "draw two cards" and then the energy: the energy waits
// for a prompt the draws ask (a dredge offer; CR 608.2c, #4155).
var glimmerOfGeniusRest = RegisterThenSteps("then-steps/glimmer-of-genius-rest",
	Instruction(func(ctx *Context) error { return ctx.Game.DrawNForEffect(ctx.Controller(), 2) }),
	Instruction(func(ctx *Context) error {
		return ctx.Game.AddPlayerCounterByForEffect(ctx.Controller(), ctx.Controller(), game.CounterEnergy, 2)
	}),
)
