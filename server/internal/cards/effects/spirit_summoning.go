package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spirit Summoning — Sorcery — Lesson {1}{R/W}{R/W} (Strixhaven: School of
// Mages, #5502):
//
//	"Create a 3/2 red and white Spirit creature token."
//
// Quintorius Kand's Spirit from the shared token table. Being a Lesson is
// its type line, which learn reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a4825093-ea29-47d9-b52c-1c7735f1909e",
		Name:         "Spirit Summoning",
		Purpose:      game.Purpose{Tokens: 1},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("3/2 red and white Spirit"), N: 1}.Apply(ctx)
		},
	})
}
