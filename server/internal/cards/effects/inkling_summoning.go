package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inkling Summoning — Sorcery — Lesson {1}{W/B}{W/B} (Strixhaven: School
// of Mages, #5502):
//
//	"Create a 2/1 white and black Inkling creature token with flying."
//
// Intermediate Chirography's Inkling from the shared token table. Being a
// Lesson is its type line, which learn reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "999a9310-d145-4c1a-8518-9d96602ca672",
		Name:         "Inkling Summoning",
		Purpose:      game.Purpose{Tokens: 1},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("2/1 white and black Inkling with flying"), N: 1}.Apply(ctx)
		},
	})
}
