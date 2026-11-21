package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Into the Story — Instant {5}{U}{U}:
//
//	"This spell costs {3} less to cast if an opponent has seven or more
//	 cards in their graveyard.
//	 Draw four cards."
//
// #746: a conditional self cost modifier.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f290c2e4-ab52-44d4-bdeb-31aeb835b18d",
		Name:         "Into the Story",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(3, "This spell costs {3} less to cast if an opponent has seven or more cards in their graveyard.",
				effects.AnOpponentHasCardsInGraveyardAtLeast(7)),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 4}.Apply(ctx)
		},
	})
}
