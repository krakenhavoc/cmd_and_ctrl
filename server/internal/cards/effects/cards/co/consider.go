package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Consider — Instant {U} (EDHREC rank 405):
//
//	"Surveil 1. (Look at the top card of your library. You may put it
//	 into your graveyard.)
//	 Draw a card."
//
// Surveil queues a prompt and nothing moves until the caster answers
// it, so the draw rides Surveil's Then rather than the next
// statement — the same ordering rule Preordain's Scry-then-draw
// follows: writing the draw as a second primitive would pull a card
// before the surveil decision is made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c9bcba6-87b5-4fb3-97ee-6fe5b739337d",
		Name:         "Consider",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			c := ctx.Controller()
			// #4253 slice 3: the draw is a registered tail, so the
			// surveil prompt is a restore point.
			return effects.Surveil{Player: c, N: 1, ThenTail: game.KeyedTail{Ref: effects.ThenDrawACard, Player: c}}.Apply(ctx)
		},
	})
}
