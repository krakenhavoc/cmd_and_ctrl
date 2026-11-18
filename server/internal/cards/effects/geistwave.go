package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "14d5ed90-0aa7-48f0-9017-83912b9e89d1",
		Name:         "Geistwave",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Not(Land())),
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := battlefieldCardForEffect(ctx.Game, id)
			if !ok {
				return nil
			}
			b := BounceToHand{Target: id}
			if c.Controller == ctx.Controller() {
				b.ThenTail = ctx.TailN(thenDrawCards, 1)
			}
			return b.Apply(ctx)
		},
	})
}
