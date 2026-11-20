package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bone to Ash — Instant {2}{U}{U} (Dark Ascension, #5571):
//
//	"Counter target creature spell.
//	 Draw a card."
//
// Dismiss narrowed to a creature spell. A target that left the stack in
// response counters the whole spell on resolution (CR 608.2b), so no card
// is drawn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "31be5bf4-9950-456c-8365-74b48e132ef5",
		Name:         "Bone to Ash",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      TargetSpell("target creature spell", Creature()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return counterTheTargetSpell(ctx.Item, ctx) }),
			DrawCards{N: 1},
		),
	})
}
