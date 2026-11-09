package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Release the Dogs — Sorcery {3}{W} (Foundations, #4465):
//
//	"Create four 1/1 white Dog creature tokens."
//
// The Dog row is the shared "1/1 white Dog" token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b97d21b4-d730-479f-872f-3e1645b66751",
		Name:         "Release the Dogs",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 4},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white Dog"), N: 4}.Apply(ctx)
		},
	})
}
