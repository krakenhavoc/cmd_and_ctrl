package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glimmer of Genius — Instant {3}{U}:
//
//	"Scry 2, then draw two cards. You get {E}{E} (two energy counters)."
//
// ADR 0129 PR 1. The draw and the energy follow the scry prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9c58cfb-b7d5-4e86-a5d3-efb07192aa22",
		Name:         "Glimmer of Genius",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Energy: 2, Scry: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.Scry{Player: ctx.Controller(), N: 2, ThenTail: ctx.ItemTail(glimmerOfGeniusThen)}.Apply(ctx)
		},
	})
}

// glimmerOfGeniusThen is the draws and the energy, once the scry is done
// (#4253 slice 4). An on-disk key.
var glimmerOfGeniusThen = effects.CardTail("glimmer-of-genius/then", func(ctx *effects.Context, _ game.TailOutcome) error {
	return glimmerOfGeniusRest.Run(ctx)
})

// glimmerOfGeniusRest is "draw two cards" and then the energy: the energy waits
// for a prompt the draws ask (a dredge offer; CR 608.2c, #4155).
var glimmerOfGeniusRest = effects.RegisterThenSteps("then-steps/glimmer-of-genius-rest",
	effects.Instruction(func(ctx *effects.Context) error { return ctx.Game.DrawNForEffect(ctx.Controller(), 2) }),
	effects.Instruction(func(ctx *effects.Context) error {
		return ctx.Game.AddPlayerCounterByForEffect(ctx.Controller(), ctx.Controller(), game.CounterEnergy, 2)
	}),
)
