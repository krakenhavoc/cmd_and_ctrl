package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Everdream — Instant {1}{U}:
//
//	"Draw a card.
//	Splice onto instant or sorcery {2}{U}"
func init() {
	Register(Spec{
		OracleID:     "5b17d3c5-6085-416e-9afd-8ed0639ae7d2",
		Name:         "Everdream",
		Completeness: CompletenessFull,
		Splice:       SpliceOntoInstantOrSorcery("{2}{U}"),
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
