package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "bb377cf8-29ee-489e-86a9-93f4dc2ae02e",
		Name:          "Field Research",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Draws: 2, Kicked: &game.Purpose{Draws: 1}},
		OptionalCosts: []game.AdditionalCost{Kicker("{2}{U}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := 2
			if ctx.WasKicked() {
				n = 3
			}
			return DrawCards{Player: ctx.Controller(), N: n}.Apply(ctx)
		},
	})
}
