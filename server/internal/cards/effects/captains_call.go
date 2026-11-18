package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain's Call — Sorcery {3}{W}:
//
//	"Create three 1/1 white Soldier creature tokens."
//
// Raise the Alarm's token clause at sorcery speed, three of them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46418fe4-065c-4dfa-b796-eee02c14f351",
		Name:         "Captain's Call",
		Purpose:      game.Purpose{Tokens: 3},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{
				Controller: ctx.Controller(),
				Template:   TokenCard("1/1 white Soldier"),
				N:          3,
			}.Apply(ctx)
		},
	})
}
