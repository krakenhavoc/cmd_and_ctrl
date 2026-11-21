package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infectious Inquiry — Sorcery {2}{B}: "You draw two cards and you lose 2
// life. Each opponent gets a poison counter." The life loss waits for a prompt
// the draw asks (CR 608.2c). No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66051196-53ea-4461-b241-205df9395b3b",
		Name:         "Infectious Inquiry",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.GainLife{Player: ctx.Controller(), Amount: -2}.Apply(ctx)
			}),
			effects.S82OneB04PoisonStep(),
		),
	})
}
