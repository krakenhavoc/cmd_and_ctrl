package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arrester's Admonition — Instant {2}{U} (Ravnica Allegiance):
//
//	"Return target creature to its owner's hand.
//	 Addendum — If you cast this spell during your main phase, draw a card."
//
// The draw rides the bounce's keyed tail, so a commander's command-zone
// prompt settles first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ccfb612a-762d-41a7-bc3d-268c6abf8e85",
		Name:         "Arrester's Admonition",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.BounceToHand{Target: id, ThenTail: ctx.Tail(arrestersAdmonitionDraw)}.Apply(ctx)
		},
	})
}

// arrestersAdmonitionDraw is the addendum's draw. An on-disk key: never
// renamed, never reused.
var arrestersAdmonitionDraw = effects.CardTail("then/arresters-admonition-draw", func(ctx *effects.Context, _ game.TailOutcome) error {
	if !effects.B14InYourMainPhase(ctx.Game, ctx.Controller()) {
		return nil
	}
	return effects.DrawCards{N: 1}.Apply(ctx)
})
