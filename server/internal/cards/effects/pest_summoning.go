package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pest Summoning — Sorcery — Lesson {1}{B/G}{B/G} (Strixhaven: School of
// Mages, #5502):
//
//	"Create two 1/1 black and green Pest creature tokens with "When this
//	 token dies, you gain 1 life.""
//
// Two of the Strixhaven Pest (PestToken), each with its own dies trigger.
// Being a Lesson is its type line, which learn reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f6e0cb77-4115-4948-8e1e-fdb774b1b254",
		Name:         "Pest Summoning",
		Purpose:      game.Purpose{Tokens: 2},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: PestToken(), N: 2}.Apply(ctx)
		},
	})
}
