package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Runeboggle — Instant {2}{U} (Guildpact, #5599):
//
//	"Counter target spell unless its controller pays {1}.
//	 Draw a card."
//
// Force Spike's question (the shared CounterUnlessPaid prompt, through
// counterTheTargetSpellUnlessPaid), then the draw. The two are
// ResolveSteps, so the draw waits for the payer's answer (CR 608.2c).
// A target spell that left the stack makes the whole spell do nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22000392-99e9-4927-a488-1bcb576f6275",
		Name:         "Runeboggle",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			return effects.CounterTheTargetSpellUnlessPaid("Runeboggle", "{1}")(ctx.Game, ctx.Item)
		}), effects.DrawCards{N: 1}),
	})
}
