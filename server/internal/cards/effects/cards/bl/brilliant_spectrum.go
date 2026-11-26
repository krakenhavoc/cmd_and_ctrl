package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brilliant Spectrum — Sorcery {3}{U} (Battle for Zendikar):
//
//	"Converge — Draw X cards, where X is the number of colors of mana spent
//	 to cast this spell. Then discard two cards."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "88ed66fc-7068-44a7-b00f-3be22fb500cb",
		Name:                "Brilliant Spectrum",
		Completeness:        effects.CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		Purpose:             game.Purpose{Draws: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DrawCards{N: ctx.ColorsSpentCount()}.Apply(ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: ctx.Controller(), Source: ctx.Source(), N: 2})
				return nil
			}),
		),
	})
}
