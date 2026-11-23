package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ral's Reinforcements — Sorcery {1}{R} (March of the Machine):
//
//	"Create two 1/1 blue and red Elemental creature tokens."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7fa2fc2-8542-4069-b75f-5520dd30f74d",
		Name:         "Ral's Reinforcements",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: ctx.TextToken(effects.TokenCard("1/1 blue and red Elemental")), N: 2}.Apply(ctx)
		},
	})
}
