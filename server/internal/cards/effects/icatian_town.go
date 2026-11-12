package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Icatian Town — Sorcery {5}{W} (Fallen Empires, #5637):
//
//	"Create four 1/1 white Citizen creature tokens."
//
// Release the Dogs with Citizens: the "1/1 white Citizen" row in the
// token table, four of them under the caster's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa8b60b6-cf55-4aaa-9caa-2b17942d8269",
		Name:         "Icatian Town",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 4},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white Citizen"), N: 4}.Apply(ctx)
		},
	})
}
