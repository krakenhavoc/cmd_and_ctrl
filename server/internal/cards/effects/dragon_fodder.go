package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragon Fodder — Sorcery {1}{R}:
//
//	"Create two 1/1 red Goblin creature tokens."
//
// Two bodies for two mana with no rider — the cheapest way an
// aristocrats deck turns a card into sacrifice fuel, and the reason
// it shows up in every Goblin and every token list.
//
// No simplification: the tokens are Krenko's, so they count for
// Krenko, Mob Boss's own ability and for anything else that counts
// Goblins.
func init() {
	Register(Spec{
		OracleID:     "d0d2c45b-b6e3-4999-bdab-976e8f0d6617",
		Name:         "Dragon Fodder",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{
				Controller: ctx.Controller(),
				Template:   RedGoblinToken(),
				N:          2,
			}.Apply(ctx)
		},
	})
}
