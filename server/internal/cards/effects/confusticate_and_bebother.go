package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7a7d9887-3cc4-4fb9-9109-1315903ace45",
		Name:         "Confusticate and Bebother",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Counter target spell unless its controller pays {4}.",
				TargetSpell("target spell"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return CounterUnlessPaid{
						StackID:  t.ID,
						Cost:     "{4}",
						Question: "Confusticate and Bebother — pay {4} or your spell is countered",
					}.Apply(ctx)
				}),
			ModeWithPurpose(ModeDoing("Draw two cards, then discard a card.", nil, drawTwoThenDiscardOne),
				game.Purpose{Draws: 2, Discards: 1}),
		),
	})
}
