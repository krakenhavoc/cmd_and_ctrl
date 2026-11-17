package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ral's Reinforcements — Sorcery {1}{R} (March of the Machine):
//
//	"Create two 1/1 blue and red Elemental creature tokens."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e7fa2fc2-8542-4069-b75f-5520dd30f74d",
		Name:         "Ral's Reinforcements",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: ctx.TextToken(TokenCard("1/1 blue and red Elemental")), N: 2}.Apply(ctx)
		},
	})
}
