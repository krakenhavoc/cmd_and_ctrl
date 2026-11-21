package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Field Research — Sorcery {2}{U} (Zendikar Rising, #5508):
//
//	"Kicker {2}{U} (You may pay an additional {2}{U} as you cast this
//	 spell.)
//	 Draw two cards. If this spell was kicked, draw three cards instead."
//
// Burst Lightning's kicker read (ctx.WasKicked, off the cast) choosing
// between the two draw counts. The kicked purpose adds the third card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "bb377cf8-29ee-489e-86a9-93f4dc2ae02e",
		Name:          "Field Research",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Draws: 2, Kicked: &game.Purpose{Draws: 1}},
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}{U}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := 2
			if ctx.WasKicked() {
				n = 3
			}
			return effects.DrawCards{Player: ctx.Controller(), N: n}.Apply(ctx)
		},
	})
}
