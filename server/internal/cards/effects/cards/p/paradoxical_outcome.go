package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paradoxical Outcome — Instant {3}{U}:
//
//	"Return any number of target nonland, nontoken permanents you control to
//	 their owners' hands. Draw a card for each card returned to your hand this
//	 way."
//
// Any number of targets, zero included (CR 115.1). Each still-legal target is
// returned in turn (CR 608.2b); the draw counts only the cards that reached
// the caster's own hand, so a permanent they control but do not own goes to
// its owner and draws nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2cce7250-f8f9-4c0a-87b1-ba3c5b8ed6b3",
		Name:         "Paradoxical Outcome",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets: effects.TargetPermanent("any number of target nonland, nontoken permanents you control",
			effects.YouControl(), effects.Nonland(), effects.Not(effects.IsTokenPredicate())).WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.BounceCardsToHandKeyedForEffect(effects.LegalTargetIDs(ctx), ctx.ItemTail(paradoxicalOutcomeDraw))
		},
	})
}

// paradoxicalOutcomeDraw is "draw a card for each card returned to your hand
// this way", keyed: the cards that reached a hand, and of those only the ones
// the caster owns (a permanent they merely control went to its owner's hand).
// The key is an on-disk identity.
var paradoxicalOutcomeDraw = effects.CardTail("then/paradoxical-outcome-draw", func(ctx *effects.Context, o game.TailOutcome) error {
	n := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.Owner == ctx.Controller() {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return effects.DrawCards{Player: ctx.Controller(), N: n}.Apply(ctx)
})
