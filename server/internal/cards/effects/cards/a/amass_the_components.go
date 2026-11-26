package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Amass the Components — {3}{U} Sorcery: "Draw three cards, then put a card from
// your hand on the bottom of your library."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f623e23-81ae-4c35-b68b-e0a2f431cbad",
		Name:         "Amass the Components",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{N: 3}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.RfPutCardFromHandOnBottom(ctx, "Amass the Components — put a card from your hand on the bottom of your library")
			}),
		),
	})
}
