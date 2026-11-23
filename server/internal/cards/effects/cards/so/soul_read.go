package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "7cf93170-29ca-4f44-a8ce-325d034d4ce8",
		Name:         "Soul Read",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Counter target spell unless its controller pays {4}.", effects.TargetSpell("target spell"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.CounterUnlessPaid{
						StackID:  t.ID,
						Cost:     "{4}",
						Question: "Soul Read — pay {4} or your spell is countered",
					}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeWithPurpose(effects.ModeDoing("Draw two cards.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DrawCards{N: 2}.Apply(ctx)
				}), game.Purpose{Draws: 2}),
		),
	})
}
