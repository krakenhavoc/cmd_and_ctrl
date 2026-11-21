package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Halt Order — Instant {2}{U} (Scars of Mirrodin, #5577):
//
//	"Counter target artifact spell.
//	 Draw a card."
//
// Artifact Blast's target on Dismiss's counter-then-draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5fdf1b2-da22-4fa6-810a-85c3f7587535",
		Name:         "Halt Order",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetSpell("target artifact spell", effects.Artifact()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.CounterTheTargetSpell(ctx.Item, ctx) }),
			effects.DrawCards{N: 1},
		),
	})
}
