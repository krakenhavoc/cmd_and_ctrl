package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gisa's Bidding — Sorcery {2}{B}{B} (Innistrad Remastered):
//
//	"Create two 2/2 black Zombie creature tokens.
//	 Madness {2}{B}"
func init() {
	Register(Spec{
		OracleID:     "0abc993b-18c7-4bb8-aa58-0d06a41cb48f",
		Name:         "Gisa's Bidding",
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Completeness: CompletenessFull,
		Madness:      "{2}{B}",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("2/2 black Zombie"), N: 2}.Apply(ctx)
		},
	})
}
