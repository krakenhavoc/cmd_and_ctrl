package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stymied Hopes — Instant {1}{U} (Theros):
//
//	"Counter target spell unless its controller pays {1}. Scry 1."
//
// Runeboggle's shape: Force Spike's question (the shared CounterUnlessPaid
// prompt), then the scry. The two are ResolveSteps, so the scry waits for the
// payer's answer (CR 608.2c). A target spell that left the stack makes the
// whole spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b081726f-0ed6-4ab0-833e-f510ac6fbeaa",
		Name:         "Stymied Hopes",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			return effects.CounterTheTargetSpellUnlessPaid("Stymied Hopes", "{1}")(ctx.Game, ctx.Item)
		}), effects.Scry{N: 1}),
	})
}
