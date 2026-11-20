package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Call the Cavalry — Sorcery {3}{W} (Dominaria, #5530):
//
//	"Create two 2/2 white Knight creature tokens with vigilance."
//
// Two 2/2 white Knight tokens with vigilance, from the token table.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "855b168b-e514-4a43-8bbe-7f1df40031ae",
		Name:         "Call the Cavalry",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: ctx.Controller(), Template: TokenCard("2/2 white Knight with vigilance"), N: 2}.Apply(ctx)
		},
	})
}
