package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unified Front — Sorcery {3}{W} (Battle for Zendikar):
//
//	"Converge — Create a 1/1 white Kor Ally creature token for each color of
//	 mana spent to cast this spell."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "1f514b02-7827-4087-a6b7-bd6a3f520a64",
		Name:                "Unified Front",
		Completeness:        effects.CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		Purpose:             game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				n := ctx.ColorsSpentCount()
				if n <= 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("1/1 white Kor Ally"), N: n}.Apply(ctx)
			}),
		),
	})
}
