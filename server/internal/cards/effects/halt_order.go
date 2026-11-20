package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Halt Order — Instant {2}{U} (Scars of Mirrodin, #5577):
//
//	"Counter target artifact spell.
//	 Draw a card."
//
// Artifact Blast's target on Dismiss's counter-then-draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a5fdf1b2-da22-4fa6-810a-85c3f7587535",
		Name:         "Halt Order",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      TargetSpell("target artifact spell", Artifact()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return counterTheTargetSpell(ctx.Item, ctx) }),
			DrawCards{N: 1},
		),
	})
}
