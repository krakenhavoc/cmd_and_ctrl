package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eagle Vision — Sorcery {4}{U} (Assassin's Creed, #5466):
//
//	"Freerunning {1}{U} (You may cast this spell for its freerunning
//	 cost if you dealt combat damage to a player this turn with an
//	 Assassin or commander.)
//	 Draw three cards."
//
// Freerunning is the alternative cost Restart Sequence uses; the draw is
// Divination's, one card more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f96f1b4c-f06a-4374-8d8f-aaa06c6108c3",
		Name:             "Eagle Vision",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Draws: 3},
		AlternativeCosts: []game.AlternativeCost{Freerunning("{1}{U}", "Assassin")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
