package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Consign to Dream — Instant {2}{U} (Shadowmoor, #5591):
//
//	"Return target permanent to its owner's hand. If that permanent is red
//	 or green, put it on top of its owner's library instead."
//
// The permanent's colours are read as the spell resolves (CR 608.2h), before
// it moves: a permanent that is both blue and red goes to the library. A
// token is returned or tucked and ceases to exist either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "624bdbfb-b611-43b6-a3c5-cc8b11dbfaae",
		Name:         "Consign to Dream",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target permanent"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			if c.HasColor(ctx.TextColor("R")) || c.HasColor(ctx.TextColor("G")) {
				return ctx.Game.TuckToLibraryForEffect(id, false)
			}
			return effects.BounceToHand{Target: id}.Apply(ctx)
		},
	})
}
