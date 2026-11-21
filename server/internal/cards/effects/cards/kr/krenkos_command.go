package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "9cfe86ae-eebe-44aa-a956-4b3e9e621105",
		Name:         "Krenko's Command",
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{
				Controller: ctx.Controller(),
				Template:   effects.RedGoblinToken(),
				N:          2,
			}.Apply(ctx)
		},
	})
}
