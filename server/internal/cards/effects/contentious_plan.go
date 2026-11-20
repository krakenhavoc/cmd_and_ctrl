package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Contentious Plan — Sorcery {1}{U} (War of the Spark, #5520):
//
//	"Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
//	 Draw a card."
//
// Steady Progress's body: the draw waits for the proliferate answer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2fc6de19-9d04-422b-90bc-eee2f9f8c84c",
		Name:         "Contentious Plan",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Proliferate{ThenTail: ctx.TailN(thenDrawCards, 1)}.Apply(ctx)
		},
	})
}
