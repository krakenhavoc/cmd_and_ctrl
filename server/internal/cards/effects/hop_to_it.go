package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hop to It — Sorcery {2}{W} (EDHREC rank 3855):
//
//	"Create three 1/1 white Rabbit creature tokens."
//
// Dragon Fodder in white, one mana dearer for a third body. It is in
// the batch for the same reason Dragon Fodder was: three bodies with
// no rider is the cheapest thing a token or aristocrats deck can do
// with a card, and it costs the engine nothing it does not already
// have.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e8a9350a-07c1-47ed-8c4f-88e4b3b17545",
		Name:         "Hop to It",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{
				Controller: ctx.Controller(),
				Template:   TokenCard("1/1 white Rabbit"),
				N:          3,
			}.Apply(ctx)
		},
	})
}
