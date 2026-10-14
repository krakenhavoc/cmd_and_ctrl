package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eyes of Gitaxias — Sorcery {2}{U} (March of the Machine):
//
//	"Incubate 3.
//	 Draw a card."
//
// Two instructions in printed order (CR 608.2c): the Incubator
// (CR 701.53a, ADR 0185) first, then the draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ef6ca6e-a8b8-4ea6-b13d-a08eb847c12d",
		Name:         "Eyes of Gitaxias",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (Incubate{N: 3}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
