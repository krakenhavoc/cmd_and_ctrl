package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1a61929e-14bd-4f21-b173-00818d49743d",
		Name:         "Memory Drain",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 2},
		Targets:      TargetSpell("target spell"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return counterTheTargetSpell(ctx.Item, ctx) }),
			Scry{N: 2},
		),
	})
}
