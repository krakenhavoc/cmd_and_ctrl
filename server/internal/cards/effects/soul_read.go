package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul Read — Instant {3}{U} (Foundations Jumpstart):
//
//	"Choose one —
//	 • Counter target spell unless its controller pays {4}.
//	 • Draw two cards."
//
// Convolute's soft counter on the first bullet (a target gone by
// resolution fizzles it, CR 608.2b), and a plain two-card draw, with no
// target, on the second.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7cf93170-29ca-4f44-a8ce-325d034d4ce8",
		Name:         "Soul Read",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Counter target spell unless its controller pays {4}.", TargetSpell("target spell"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return CounterUnlessPaid{
						StackID:  t.ID,
						Cost:     "{4}",
						Question: "Soul Read — pay {4} or your spell is countered",
					}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerRemove}),
			ModeWithPurpose(ModeDoing("Draw two cards.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return DrawCards{N: 2}.Apply(ctx)
				}), game.Purpose{Draws: 2}),
		),
	})
}
