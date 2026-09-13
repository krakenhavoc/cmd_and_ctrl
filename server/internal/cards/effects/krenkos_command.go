package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Krenko's Command — Sorcery {1}{R} (EDHREC rank 2033):
//
//	"Create two 1/1 red Goblin creature tokens."
//
// Dragon Fodder with a different name — the second copy every
// Goblin deck runs. The tokens are Krenko's own template, so they
// count for Krenko, Mob Boss and anything else that counts Goblins.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9cfe86ae-eebe-44aa-a956-4b3e9e621105",
		Name:         "Krenko's Command",
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
