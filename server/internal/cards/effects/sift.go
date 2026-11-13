package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sift — Sorcery {3}{U} (Stronghold, #5627):
//
//	"Draw three cards, then discard a card."
//
// Focus the Mind's shared drawThenDiscard body at three and one, without
// its discount: the three cards are drawn first, so any of them can be
// the one discarded, and the caster picks it (CR 701.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6e060984-2015-4dc4-b53d-ab6bf2abdacc",
		Name:         "Sift",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3, Discards: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 3, 1, "")
		},
	})
}
