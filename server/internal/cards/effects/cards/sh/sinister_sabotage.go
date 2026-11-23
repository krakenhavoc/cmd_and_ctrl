package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sinister Sabotage — Instant {1}{U}{U} (Guilds of Ravnica, #5525):
//
//	"Counter target spell.
//	 Surveil 1."
//
// Two steps: Cancel's counter (counterTheTargetSpell), then surveil 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "973fd4d4-9255-4825-85b7-503606c4e932",
		Name:         "Sinister Sabotage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.CounterTheTargetSpell(ctx.Item, ctx) }),
			effects.Surveil{N: 1},
		),
	})
}
