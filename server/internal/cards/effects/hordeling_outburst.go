package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hordeling Outburst — Sorcery {1}{R}{R}:
//
//	"Create three 1/1 red Goblin creature tokens."
//
// Dragon Fodder's big brother: one more mana, one more body, and the
// three-for-one that makes a sacrifice outlet worth the card. Same
// token, same lack of rider.
func init() {
	Register(Spec{
		OracleID: "a6450b8e-eb18-431c-9eb7-7daf107978b2",
		Name:     "Hordeling Outburst",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{
				Controller: ctx.Controller(),
				Template:   RedGoblinToken(),
				N:          3,
			}.Apply(ctx)
		},
	})
}
