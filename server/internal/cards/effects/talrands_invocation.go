package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Talrand's Invocation — Sorcery {2}{U}{U} (Jumpstart 2020):
//
//	"Create two 2/2 blue Drake creature tokens with flying."
//
// The Drake is the same token row Talrand, Sky Summoner makes.
//
// No simplification. Deck request #4894 (Jumpstart Azorius).
func init() {
	Register(Spec{
		OracleID:     "43355ac4-bf8c-48f6-a322-fafbc9d132d1",
		Name:         "Talrand's Invocation",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: item.Controller, Template: TokenCard("2/2 blue Drake with flying"), N: 2}.Apply(ctx)
		},
	})
}
