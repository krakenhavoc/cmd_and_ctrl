package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghastly Discovery — Sorcery {2}{U} (Shadowmoor, S82 older sets):
//
//	"Draw two cards, then discard a card.
//	 Conspire (As you cast this spell, you may tap two untapped creatures you
//	 control that share a color with it. When you do, copy it.)"
//
// Snow Day's loot (the draw first, so a drawn card is a legal discard) with conspire; the copy loots again.
// Conspire is the engine's optional cost and cast trigger (conspire.go, CR
// 702.78), so the card declares the cost only. The conspire copy does the same again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "1ca875f0-50a7-4332-8643-c2b4bbda18b1",
		Name:          "Ghastly Discovery",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Draws: 2, Discards: 1},
		OptionalCosts: []game.AdditionalCost{effects.Conspire()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 2, 1, "")
		},
	})
}
