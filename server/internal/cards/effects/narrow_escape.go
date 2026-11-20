package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Narrow Escape — Instant {2}{W} (Zendikar, #5583):
//
//	"Return target permanent you control to its owner's hand. You gain
//	 4 life."
//
// Rescue's bounce, then the life. The spell has one target, so if it is
// gone by resolution the spell does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c96803dc-b69b-4fc3-a39b-1aad6d800e86",
		Name:         "Narrow Escape",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target permanent you control", YouControl()),
		Purpose:      game.Purpose{Answers: game.AnswerSurvive},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return bounceTheTarget(ctx.Item, ctx) }),
			GainLife{Amount: 4},
		),
	})
}
