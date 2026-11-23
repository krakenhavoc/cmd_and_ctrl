package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geistwave — Instant {1}{U} (Innistrad: Midnight Hunt, #5497):
//
//	"Return target nonland permanent to its owner's hand. If you controlled
//	 that permanent, draw a card."
//
// Who controlled it is read before it leaves (CR 608.2h); the draw is the
// bounce's continuation, so it waits for a commander's command-zone question.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14d5ed90-0aa7-48f0-9017-83912b9e89d1",
		Name:         "Geistwave",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Not(effects.Land())),
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := effects.BattlefieldCardForEffect(ctx.Game, id)
			if !ok {
				return nil
			}
			b := effects.BounceToHand{Target: id}
			if c.Controller == ctx.Controller() {
				b.ThenTail = ctx.TailN(effects.ThenDrawCards, 1)
			}
			return b.Apply(ctx)
		},
	})
}
