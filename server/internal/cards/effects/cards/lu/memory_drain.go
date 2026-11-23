package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Memory Drain — Instant {2}{U}{U} (Theros Beyond Death, #5514):
//
//	"Counter target spell. Scry 2."
//
// Cancel's counter, then the scry, in printed order (CR 608.2c). A
// target that has left the stack makes the spell do nothing, the scry
// included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a61929e-14bd-4f21-b173-00818d49743d",
		Name:         "Memory Drain",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 2},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.CounterTheTargetSpell(ctx.Item, ctx) }),
			effects.Scry{N: 2},
		),
	})
}
