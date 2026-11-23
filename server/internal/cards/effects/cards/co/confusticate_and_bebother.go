package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Confusticate and Bebother — Instant {2}{U} (The Hobbit, HOB):
//
//	"Choose one —
//	 • Counter target spell unless its controller pays {4}.
//	 • Draw two cards, then discard a card."
//
// The first bullet is Mystic Confluence's counter-unless-pays (the pay
// prompt goes to the spell's controller, declining counters it); the second
// is Three Steps Ahead's loot (drawTwoThenDiscardOne), the draw first so a
// drawn card is a legal discard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a7d9887-3cc4-4fb9-9109-1315903ace45",
		Name:         "Confusticate and Bebother",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Counter target spell unless its controller pays {4}.",
				effects.TargetSpell("target spell"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.CounterUnlessPaid{
						StackID:  t.ID,
						Cost:     "{4}",
						Question: "Confusticate and Bebother — pay {4} or your spell is countered",
					}.Apply(ctx)
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Draw two cards, then discard a card.", nil, effects.DrawTwoThenDiscardOne),
				game.Purpose{Draws: 2, Discards: 1}),
		),
	})
}
