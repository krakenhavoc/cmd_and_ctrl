package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sphinx of Lost Truths — Creature — Sphinx {3}{U}{U} 3/5 (Zendikar, S82 OLD
// batch 491, #5583):
//
//	"Kicker {1}{U}
//	 Flying
//	 When this creature enters, draw three cards. Then if it wasn't kicked,
//	 discard three cards."
//
// The kicked count is recorded when the trigger goes on the stack
// (S82OldB491WhenThisEntersCapturingKicked), so the answer survives the Sphinx
// leaving in response. Unkicked, the draw and the discard are one
// draw-then-discard: the discard prompt opens after all three draws, so the
// drawn cards are legal discards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "25597913-aa17-4e3b-a4e9-371ce1459160",
		Name:            "Sphinx of Lost Truths",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{1}{U}")},
		Purpose:         game.Purpose{Draws: 3},
		Triggered: []game.TriggeredAbility{
			effects.S82OldB491WhenThisEntersCapturingKicked("Sphinx of Lost Truths — draw three cards; then if it wasn't kicked, discard three cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if item.Params.Amount > 0 {
						return g.DrawNForEffect(ctx.Controller(), 3)
					}
					return effects.DrawThenDiscard(g, ctx.Controller(), ctx.Source(), 3, 3, "")
				}),
		},
	})
}
