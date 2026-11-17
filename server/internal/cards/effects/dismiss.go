package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dismiss — Instant {2}{U}{U} (Tempest, #5628):
//
//	"Counter target spell.
//	 Draw a card."
//
// Counterspell's counter (counterTheTargetSpell, CR 701.6a), then the
// draw (CR 608.2c). The target is a spell only, as printed: an ability
// on the stack can't be chosen. A spell that left the stack in response
// is an illegal target, so Dismiss doesn't resolve and nothing is drawn
// (CR 608.2b), as in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5c828a5e-10ae-4f63-86fd-2f160af43cc3",
		Name:         "Dismiss",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      TargetSpell("target spell"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return counterTheTargetSpell(ctx.Item, ctx) }),
			DrawCards{N: 1},
		),
	})
}
